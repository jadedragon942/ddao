package s3

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
)

// S3Storage implements the DDAO storage interface using Amazon S3
type S3Storage struct {
	client    *s3.Client
	uploader  *manager.Uploader
	bucket    string
	prefix    string
	region    string
	sch       *schema.Schema
	verbose   bool
}

// S3Object represents a stored object in S3
type S3Object struct {
	ID        string                 `json:"id"`
	TableName string                 `json:"table_name"`
	Fields    map[string]interface{} `json:"fields"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
}

// S3Transaction represents a transaction context for S3 operations
type S3Transaction struct {
	ID        string
	CreatedAt time.Time
	Objects   map[string]*S3Object // key format: "table/id"
}

func New() storage.Storage {
	return &S3Storage{}
}

// Connect establishes a connection to S3
// connStr format: "s3://bucket/prefix?region=us-east-1&endpoint=http://localhost:9000"
// Examples:
//   - "s3://my-bucket/ddao-data?region=us-east-1"
//   - "s3://my-bucket/ddao-data?region=us-east-1&endpoint=http://localhost:9000" (for MinIO)
func (s *S3Storage) Connect(ctx context.Context, connStr string) error {
	// Parse connection string
	u, err := url.Parse(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	if u.Scheme != "s3" {
		return fmt.Errorf("invalid scheme: expected s3, got %s", u.Scheme)
	}

	s.bucket = u.Host
	s.prefix = strings.TrimPrefix(u.Path, "/")
	if s.prefix != "" && !strings.HasSuffix(s.prefix, "/") {
		s.prefix += "/"
	}

	query := u.Query()
	s.region = query.Get("region")
	if s.region == "" {
		s.region = "us-east-1" // default region
	}

	endpoint := query.Get("endpoint")
	s.verbose = query.Get("verbose") == "true"

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(s.region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	var options []func(*s3.Options)
	if endpoint != "" {
		options = append(options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // Required for MinIO and some S3-compatible services
		})
	}

	s.client = s3.NewFromConfig(cfg, options...)
	s.uploader = manager.NewUploader(s.client)

	// Test connection by checking if bucket exists
	storage.DebugLog("HeadBucket", s.bucket)
	_, err = s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to access bucket %s: %w", s.bucket, err)
	}

	if s.verbose {
		log.Printf("Connected to S3 bucket: %s, prefix: %s, region: %s", s.bucket, s.prefix, s.region)
	}

	return nil
}

// CreateTables creates the necessary "table" structure in S3
// For S3, this means creating metadata files for each table
func (s *S3Storage) CreateTables(ctx context.Context, schema *schema.Schema) error {
	if s.client == nil {
		return errors.New("not connected to S3")
	}

	s.sch = schema

	// Create a metadata file for the schema
	schemaData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	schemaKey := s.prefix + "_schema.json"
	storage.DebugLog("PutObject (schema)", schemaKey)
	_, err = s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(schemaKey),
		Body:   bytes.NewReader(schemaData),
		Metadata: map[string]string{
			"ddao-type":      "schema",
			"ddao-timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload schema: %w", err)
	}

	// Create metadata files for each table
	for _, table := range schema.Tables {
		tableMetadata := map[string]interface{}{
			"table_name": table.TableName,
			"fields":     table.Fields,
			"created_at": time.Now().UTC(),
		}

		metadataData, err := json.MarshalIndent(tableMetadata, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal table metadata for %s: %w", table.TableName, err)
		}

		metadataKey := s.prefix + "tables/" + table.TableName + "/_metadata.json"
		storage.DebugLog("PutObject (table metadata)", metadataKey)
		_, err = s.uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(metadataKey),
			Body:   bytes.NewReader(metadataData),
			Metadata: map[string]string{
				"ddao-type":      "table-metadata",
				"ddao-table":     table.TableName,
				"ddao-timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to upload table metadata for %s: %w", table.TableName, err)
		}

		if s.verbose {
			log.Printf("Created table metadata for: %s", table.TableName)
		}
	}

	if s.verbose {
		log.Printf("Created %d tables in S3", len(schema.Tables))
	}

	return nil
}

// Insert creates a new object in S3
func (s *S3Storage) Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {

	if s.client == nil {
		return nil, false, errors.New("not connected to S3")
	}

	// Check if object already exists
	objectKey := s.getObjectKey(obj.TableName, obj.ID)
	storage.DebugLog("GetObject (check exists)", objectKey)
	_, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})

	created := false
	if err != nil {
		// Object doesn't exist, we can create it
		var nfe *types.NoSuchKey
		if errors.As(err, &nfe) {
			created = true
		} else {
			return nil, false, fmt.Errorf("failed to check if object exists: %w", err)
		}
	}

	// Create S3Object
	s3Obj := &S3Object{
		ID:        obj.ID,
		TableName: obj.TableName,
		Fields:    obj.Fields,
		CreatedAt: time.Now().UTC(),
	}

	if !created {
		// If updating existing object, set UpdatedAt
		s3Obj.UpdatedAt = time.Now().UTC()
	}

	// Serialize object
	objData, err := json.MarshalIndent(s3Obj, "", "  ")
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal object: %w", err)
	}

	// Upload to S3
	storage.DebugLog("PutObject (insert)", objectKey)
	_, err = s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(objData),
		Metadata: map[string]string{
			"ddao-type":      "object",
			"ddao-table":     obj.TableName,
			"ddao-id":        obj.ID,
			"ddao-timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to upload object: %w", err)
	}

	if s.verbose {
		log.Printf("Inserted object: %s/%s (created: %v)", obj.TableName, obj.ID, created)
	}

	return objData, created, nil
}

// Update modifies an existing object in S3
func (s *S3Storage) Update(ctx context.Context, obj *object.Object) (bool, error) {

	if s.client == nil {
		return false, errors.New("not connected to S3")
	}

	objectKey := s.getObjectKey(obj.TableName, obj.ID)

	// Check if object exists
	storage.DebugLog("GetObject (update check)", objectKey)
	_, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var nfe *types.NoSuchKey
		if errors.As(err, &nfe) {
			return false, nil // Object doesn't exist
		}
		return false, fmt.Errorf("failed to check if object exists: %w", err)
	}

	// Create updated S3Object
	s3Obj := &S3Object{
		ID:        obj.ID,
		TableName: obj.TableName,
		Fields:    obj.Fields,
		UpdatedAt: time.Now().UTC(),
	}

	// Serialize object
	objData, err := json.MarshalIndent(s3Obj, "", "  ")
	if err != nil {
		return false, fmt.Errorf("failed to marshal object: %w", err)
	}

	// Upload to S3
	storage.DebugLog("PutObject (update)", objectKey)
	_, err = s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(objData),
		Metadata: map[string]string{
			"ddao-type":      "object",
			"ddao-table":     obj.TableName,
			"ddao-id":        obj.ID,
			"ddao-timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to upload updated object: %w", err)
	}

	if s.verbose {
		log.Printf("Updated object: %s/%s", obj.TableName, obj.ID)
	}

	return true, nil
}

// Upsert inserts or updates an object, delegating to Insert which already implements upsert behavior
func (s *S3Storage) Upsert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	return s.Insert(ctx, obj)
}

// UpsertTx inserts or updates an object within a transaction, delegating to InsertTx which already implements upsert behavior
func (s *S3Storage) UpsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	return s.InsertTx(ctx, tx, obj)
}

// FindByID retrieves an object by its ID
func (s *S3Storage) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {

	if s.client == nil {
		return nil, errors.New("not connected to S3")
	}

	objectKey := s.getObjectKey(tblName, id)

	// Get object from S3
	storage.DebugLog("GetObject (find by id)", objectKey)
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var nfe *types.NoSuchKey
		if errors.As(err, &nfe) {
			return nil, nil // Object not found
		}
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read and parse object data
	objData, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	var s3Obj S3Object
	err = json.Unmarshal(objData, &s3Obj)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal object: %w", err)
	}

	// Convert to DDAO object
	obj := &object.Object{
		ID:        s3Obj.ID,
		TableName: s3Obj.TableName,
		Fields:    s3Obj.Fields,
	}

	if s.verbose {
		log.Printf("Found object: %s/%s", tblName, id)
	}

	return obj, nil
}

// FindByKey searches for objects by a specific field value
func (s *S3Storage) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {

	if s.client == nil {
		return nil, errors.New("not connected to S3")
	}

	// List all objects in the table
	tablePrefix := s.prefix + "tables/" + tblName + "/objects/"

	storage.DebugLog("ListObjectsV2 (find by key)", tablePrefix)
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(tablePrefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			if strings.HasSuffix(*obj.Key, ".json") {
				// Get and check this object
				result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(s.bucket),
					Key:    obj.Key,
				})
				if err != nil {
					continue // Skip this object
				}

				objData, err := io.ReadAll(result.Body)
				result.Body.Close()
				if err != nil {
					continue // Skip this object
				}

				var s3Obj S3Object
				err = json.Unmarshal(objData, &s3Obj)
				if err != nil {
					continue // Skip this object
				}

				// Check if the field matches
				if fieldValue, exists := s3Obj.Fields[key]; exists {
					if fmt.Sprintf("%v", fieldValue) == value {
						// Found matching object
						ddaoObj := &object.Object{
							ID:        s3Obj.ID,
							TableName: s3Obj.TableName,
							Fields:    s3Obj.Fields,
						}

						if s.verbose {
							log.Printf("Found object by key %s=%s: %s/%s", key, value, tblName, s3Obj.ID)
						}

						return ddaoObj, nil
					}
				}
			}
		}
	}

	return nil, nil // Not found
}

// DeleteByID removes an object by its ID
func (s *S3Storage) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {

	if s.client == nil {
		return false, errors.New("not connected to S3")
	}

	objectKey := s.getObjectKey(tblName, id)

	// Check if object exists
	storage.DebugLog("GetObject (delete check)", objectKey)
	_, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var nfe *types.NoSuchKey
		if errors.As(err, &nfe) {
			return false, nil // Object doesn't exist
		}
		return false, fmt.Errorf("failed to check if object exists: %w", err)
	}

	// Delete object
	storage.DebugLog("DeleteObject", objectKey)
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete object: %w", err)
	}

	if s.verbose {
		log.Printf("Deleted object: %s/%s", tblName, id)
	}

	return true, nil
}

// ResetConnection closes the S3 connection
func (s *S3Storage) ResetConnection(ctx context.Context) error {

	s.client = nil
	s.uploader = nil

	if s.verbose {
		log.Printf("Disconnected from S3")
	}

	return nil
}

// getObjectKey returns the S3 key for a specific object
func (s *S3Storage) getObjectKey(tableName, id string) string {
	return s.prefix + "tables/" + tableName + "/objects/" + id + ".json"
}

// Transaction support - S3 doesn't have native transactions, so we implement a simple version
// Note: This is a basic implementation and not ACID compliant

func (s *S3Storage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	// S3 doesn't support traditional transactions
	// Return a mock transaction that we'll handle in our transaction methods
	return &sql.Tx{}, nil
}

func (s *S3Storage) CommitTx(tx *sql.Tx) error {
	// For S3, commit is a no-op since we don't buffer operations
	return nil
}

func (s *S3Storage) RollbackTx(tx *sql.Tx) error {
	// For S3, rollback is limited since we can't undo S3 operations easily
	// In a real implementation, you might implement a transaction log
	return nil
}

func (s *S3Storage) InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	// For simplicity, delegate to regular Insert
	return s.Insert(ctx, obj)
}

func (s *S3Storage) UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error) {
	// For simplicity, delegate to regular Update
	return s.Update(ctx, obj)
}

func (s *S3Storage) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error) {
	// For simplicity, delegate to regular FindByID
	return s.FindByID(ctx, tblName, id)
}

func (s *S3Storage) FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error) {
	// For simplicity, delegate to regular FindByKey
	return s.FindByKey(ctx, tblName, key, value)
}

func (s *S3Storage) DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error) {
	// For simplicity, delegate to regular DeleteByID
	return s.DeleteByID(ctx, tblName, id)
}