package scylla

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
)

type ScyllaDBStorage struct {
	session  *gocql.Session
	cluster  *gocql.ClusterConfig
	keyspace string
	sch      *schema.Schema
}

func New() storage.Storage {
	return &ScyllaDBStorage{}
}

// Connect establishes a connection to ScyllaDB
// connStr format: "hosts,hosts,hosts/keyspace?consistency=quorum&timeout=10s"
// Example: "localhost:9042,192.168.1.2:9042/mykeyspace?consistency=quorum&timeout=5s"
func (s *ScyllaDBStorage) Connect(ctx context.Context, connStr string) error {
	parts := strings.Split(connStr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid connection string format, expected: hosts/keyspace?options")
	}

	hosts := strings.Split(parts[0], ",")
	keyspaceAndOptions := parts[1]

	keyspaceParts := strings.Split(keyspaceAndOptions, "?")
	s.keyspace = keyspaceParts[0]

	// Create cluster configuration
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = s.keyspace
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = 10 * time.Second
	cluster.Timeout = 10 * time.Second

	// Parse connection options if provided
	if len(keyspaceParts) > 1 {
		options := keyspaceParts[1]
		err := s.parseConnectionOptions(cluster, options)
		if err != nil {
			return fmt.Errorf("failed to parse connection options: %w", err)
		}
	}

	// Create session
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create ScyllaDB session: %w", err)
	}

	s.session = session
	s.cluster = cluster

	log.Printf("Connected to ScyllaDB cluster with keyspace: %s", s.keyspace)
	return nil
}

func (s *ScyllaDBStorage) parseConnectionOptions(cluster *gocql.ClusterConfig, options string) error {
	opts := strings.Split(options, "&")
	for _, opt := range opts {
		kv := strings.Split(opt, "=")
		if len(kv) != 2 {
			continue
		}

		key, value := kv[0], kv[1]
		switch key {
		case "consistency":
			consistency, err := s.parseConsistency(value)
			if err != nil {
				return err
			}
			cluster.Consistency = consistency
		case "timeout":
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid timeout format: %w", err)
			}
			cluster.Timeout = duration
			cluster.ConnectTimeout = duration
		}
	}
	return nil
}

func (s *ScyllaDBStorage) parseConsistency(value string) (gocql.Consistency, error) {
	switch strings.ToLower(value) {
	case "any":
		return gocql.Any, nil
	case "one":
		return gocql.One, nil
	case "two":
		return gocql.Two, nil
	case "three":
		return gocql.Three, nil
	case "quorum":
		return gocql.Quorum, nil
	case "all":
		return gocql.All, nil
	case "localone":
		return gocql.LocalOne, nil
	case "localquorum":
		return gocql.LocalQuorum, nil
	case "eachquorum":
		return gocql.EachQuorum, nil
	default:
		return gocql.Quorum, fmt.Errorf("unsupported consistency level: %s", value)
	}
}

func (s *ScyllaDBStorage) CreateTables(ctx context.Context, schema *schema.Schema) error {
	if s.session == nil {
		return errors.New("not connected")
	}

	// Create keyspace if it doesn't exist
	createKeyspaceQuery := fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 3}",
		s.keyspace)

	storage.DebugLog(createKeyspaceQuery)
	if err := s.session.Query(createKeyspaceQuery).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace %s: %w", s.keyspace, err)
	}

	for _, table := range schema.Tables {
		createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text PRIMARY KEY", s.keyspace, table.TableName)

		for _, field := range table.Fields {
			if field.Name == "id" {
				continue // Skip the id field, it's already handled
			}

			// Map data types to ScyllaDB/Cassandra equivalents
			scyllaDataType := s.mapDataType(field.DataType)
			createTableQuery += fmt.Sprintf(", %s %s", field.Name, scyllaDataType)
		}

		createTableQuery += ")"

		storage.DebugLog(createTableQuery)
		log.Printf("Creating table %s with query: %s", table.TableName, createTableQuery)

		if err := s.session.Query(createTableQuery).Exec(); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.TableName, err)
		}
	}

	s.sch = schema
	log.Println("Tables created successfully")

	return nil
}

func (s *ScyllaDBStorage) mapDataType(dataType string) string {
	switch strings.ToUpper(dataType) {
	case "TEXT", "VARCHAR", "CHAR":
		return "text"
	case "INTEGER", "INT":
		return "bigint"
	case "REAL", "FLOAT":
		return "double"
	case "BLOB":
		return "blob"
	case "BOOLEAN":
		return "boolean"
	case "JSON":
		return "text" // ScyllaDB doesn't have native JSON type, store as text
	case "DATETIME", "TIMESTAMP":
		return "timestamp"
	case "DATE":
		return "date"
	case "TIME":
		return "time"
	case "UUID":
		return "uuid"
	default:
		return "text"
	}
}

func (s *ScyllaDBStorage) Insert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	if s.session == nil {
		return nil, false, errors.New("not connected")
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, false, err
	}

	if s.sch == nil {
		return nil, false, errors.New("schema not initialized")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(obj.TableName)
	if !ok {
		return nil, false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	columns := make([]string, 0, len(obj.Fields)+1)
	placeholders := make([]string, 0, len(obj.Fields)+1)
	values := make([]interface{}, 0, len(obj.Fields)+1)

	columns = append(columns, "id")
	placeholders = append(placeholders, "?")
	values = append(values, obj.ID)

	for name, field := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}

		columns = append(columns, name)
		placeholders = append(placeholders, "?")

		schField, ok := tbl.Fields[name]
		if !ok {
			return nil, false, fmt.Errorf("field %s not found in table %s schema", name, tbl.TableName)
		}
		if strings.ToLower(schField.DataType) == "json" {
			jsonData, err := json.Marshal(field)
			if err != nil {
				return nil, false, fmt.Errorf("failed to marshal JSON field %s: %w", name, err)
			}
			values = append(values, string(jsonData))
		} else {
			values = append(values, field)
		}
	}

	query := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		s.keyspace,
		tbl.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	storage.DebugLog(query, values...)

	if err := s.session.Query(query, values...).Exec(); err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (s *ScyllaDBStorage) Update(ctx context.Context, obj *object.Object) (bool, error) {
	if s.session == nil {
		return false, errors.New("not connected")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(obj.TableName)
	if !ok {
		return false, fmt.Errorf("table %s not found in schema", obj.TableName)
	}

	setClauses := make([]string, 0, len(obj.Fields)-1)
	values := make([]interface{}, 0, len(obj.Fields))

	for name, value := range obj.Fields {
		if strings.ToLower(name) == "id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", name))
		values = append(values, value)
	}

	values = append(values, obj.ID)

	query := fmt.Sprintf("UPDATE %s.%s SET %s WHERE id = ?",
		s.keyspace, tbl.TableName, strings.Join(setClauses, ", "))

	storage.DebugLog(query, values...)

	if err := s.session.Query(query, values...).Exec(); err != nil {
		return false, err
	}

	// ScyllaDB doesn't return affected rows count in the same way as SQL databases
	// We assume the update was successful if no error occurred
	return true, nil
}

// Upsert inserts or updates an object, delegating to Insert which already implements upsert behavior using INSERT INTO
func (s *ScyllaDBStorage) Upsert(ctx context.Context, obj *object.Object) ([]byte, bool, error) {
	return s.Insert(ctx, obj)
}

// UpsertTx inserts or updates an object within a transaction, delegating to InsertTx which returns an error since ScyllaDB doesn't support SQL-style transactions
func (s *ScyllaDBStorage) UpsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	return s.InsertTx(ctx, tx, obj)
}

func (s *ScyllaDBStorage) FindByID(ctx context.Context, tblName, id string) (*object.Object, error) {
	return s.FindByKey(ctx, tblName, "id", id)
}

func (s *ScyllaDBStorage) FindByKey(ctx context.Context, tblName, key, value string) (*object.Object, error) {
	if tblName == "" || key == "" || value == "" {
		return nil, errors.New("table name, key, and value must not be empty")
	}

	if s.session == nil {
		return nil, errors.New("not connected")
	}

	s.sch.Lock()
	defer s.sch.Unlock()

	tbl, ok := s.sch.GetTable(tblName)
	if !ok {
		return nil, fmt.Errorf("table %s not found in schema", tblName)
	}

	// Create ordered slices of fields to ensure consistent mapping
	fieldNames := make([]string, 0, len(tbl.Fields))
	for _, field := range tbl.Fields {
		fieldNames = append(fieldNames, field.Name)
	}

	columns := make([]string, 0, len(fieldNames))
	columnPointers := make([]interface{}, 0, len(fieldNames))

	for _, fieldName := range fieldNames {
		field := tbl.Fields[fieldName]
		columns = append(columns, field.Name)

		log.Printf("Processing field: %s with data type: %s", field.Name, field.DataType)
		switch strings.ToUpper(field.DataType) {
		case "TEXT", "VARCHAR", "CHAR":
			columnPointers = append(columnPointers, new(string))
		case "INTEGER", "INT":
			columnPointers = append(columnPointers, new(int64))
		case "REAL", "FLOAT":
			columnPointers = append(columnPointers, new(float64))
		case "BLOB":
			columnPointers = append(columnPointers, new([]byte))
		case "BOOLEAN":
			columnPointers = append(columnPointers, new(bool))
		case "JSON":
			columnPointers = append(columnPointers, new(string))
		case "DATETIME", "TIMESTAMP":
			columnPointers = append(columnPointers, new(time.Time))
		case "DATE":
			columnPointers = append(columnPointers, new(time.Time))
		case "TIME":
			columnPointers = append(columnPointers, new(time.Time))
		case "UUID":
			columnPointers = append(columnPointers, new(gocql.UUID))
		default:
			columnPointers = append(columnPointers, new(string))
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s.%s WHERE %s = ?",
		strings.Join(columns, ", "), s.keyspace, tbl.TableName, key)

	storage.DebugLog(query, value)

	iter := s.session.Query(query, value).Iter()
	defer iter.Close()

	if !iter.Scan(columnPointers...) {
		if err := iter.Close(); err != nil {
			return nil, err
		}
		return nil, nil // No rows found
	}

	// Create the object and populate fields after scanning
	var obj object.Object
	obj.TableName = tbl.TableName
	obj.Fields = make(map[string]interface{})

	for i, fieldName := range fieldNames {
		field := tbl.Fields[fieldName]
		switch strings.ToUpper(field.DataType) {
		case "TEXT", "VARCHAR", "CHAR":
			obj.Fields[field.Name] = *columnPointers[i].(*string)
		case "INTEGER", "INT":
			obj.Fields[field.Name] = *columnPointers[i].(*int64)
		case "REAL", "FLOAT":
			obj.Fields[field.Name] = *columnPointers[i].(*float64)
		case "BLOB":
			obj.Fields[field.Name] = *columnPointers[i].(*[]byte)
		case "BOOLEAN":
			obj.Fields[field.Name] = *columnPointers[i].(*bool)
		case "JSON":
			obj.Fields[field.Name] = *columnPointers[i].(*string)
		case "DATETIME", "TIMESTAMP", "DATE", "TIME":
			obj.Fields[field.Name] = *columnPointers[i].(*time.Time)
		case "UUID":
			uuid := *columnPointers[i].(*gocql.UUID)
			obj.Fields[field.Name] = uuid.String()
		default:
			obj.Fields[field.Name] = *columnPointers[i].(*string)
		}
	}

	// Set the ID field from the object fields
	if idValue, exists := obj.Fields["id"]; exists && idValue != nil {
		obj.ID = idValue.(string)
	}

	return &obj, nil
}

func (s *ScyllaDBStorage) DeleteByID(ctx context.Context, tblName, id string) (bool, error) {
	if s.session == nil {
		return false, errors.New("not connected")
	}

	query := fmt.Sprintf("DELETE FROM %s.%s WHERE id = ?", s.keyspace, tblName)

	storage.DebugLog(query, id)
	if err := s.session.Query(query, id).Exec(); err != nil {
		return false, err
	}

	// ScyllaDB doesn't return affected rows count in the same way as SQL databases
	// We assume the delete was successful if no error occurred
	return true, nil
}

func (s *ScyllaDBStorage) ResetConnection(ctx context.Context) error {
	if s.session != nil {
		s.session.Close()
		s.session = nil
	}
	return nil
}

// Transaction support methods - ScyllaDB doesn't support traditional ACID transactions
// These methods return errors indicating that transactions are not supported
func (s *ScyllaDBStorage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return nil, errors.New("ScyllaDB does not support SQL-style transactions")
}

func (s *ScyllaDBStorage) CommitTx(tx *sql.Tx) error {
	return errors.New("ScyllaDB does not support SQL-style transactions")
}

func (s *ScyllaDBStorage) RollbackTx(tx *sql.Tx) error {
	return errors.New("ScyllaDB does not support SQL-style transactions")
}

func (s *ScyllaDBStorage) InsertTx(ctx context.Context, tx *sql.Tx, obj *object.Object) ([]byte, bool, error) {
	return nil, false, errors.New("ScyllaDB does not support SQL-style transactions")
}

func (s *ScyllaDBStorage) UpdateTx(ctx context.Context, tx *sql.Tx, obj *object.Object) (bool, error) {
	return false, errors.New("ScyllaDB does not support SQL-style transactions")
}

func (s *ScyllaDBStorage) FindByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (*object.Object, error) {
	return nil, errors.New("ScyllaDB does not support SQL-style transactions")
}

func (s *ScyllaDBStorage) FindByKeyTx(ctx context.Context, tx *sql.Tx, tblName, key, value string) (*object.Object, error) {
	return nil, errors.New("ScyllaDB does not support SQL-style transactions")
}

func (s *ScyllaDBStorage) DeleteByIDTx(ctx context.Context, tx *sql.Tx, tblName, id string) (bool, error) {
	return false, errors.New("ScyllaDB does not support SQL-style transactions")
}
