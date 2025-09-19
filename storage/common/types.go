package common

import (
	"database/sql"
	"strings"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/schema"
)

// SQLStorage represents common functionality for SQL-based storage adapters
type SQLStorage struct {
	DB  *sql.DB
	Sch *schema.Schema
}

// FieldScanner handles common field scanning logic for different data types
type FieldScanner struct {
	ColumnPointers []any
	FieldTypes     []string
	FieldNames     []string
	Fields         map[string]schema.ColumnData
}

// NewFieldScanner creates a new field scanner for the given table schema
func NewFieldScanner(tbl schema.TableSchema) *FieldScanner {
	fieldNames := tbl.FieldOrder
	columnPointers := make([]any, 0, len(fieldNames))
	fieldTypes := make([]string, 0, len(fieldNames))

	for _, fieldName := range fieldNames {
		field := tbl.Fields[fieldName]
		fieldTypes = append(fieldTypes, field.DataType)

		switch strings.ToUpper(field.DataType) {
		case "TEXT", "VARCHAR", "CHAR":
			if field.Nullable {
				columnPointer := new(*string)
				columnPointers = append(columnPointers, columnPointer)
			} else {
				columnPointer := new(string)
				columnPointers = append(columnPointers, columnPointer)
			}
		case "INTEGER", "INT":
			if field.Nullable {
				columnPointer := new(*int64)
				columnPointers = append(columnPointers, columnPointer)
			} else {
				columnPointer := new(int64)
				columnPointers = append(columnPointers, columnPointer)
			}
		case "REAL", "FLOAT":
			columnPointer := new(float64)
			columnPointers = append(columnPointers, columnPointer)
		case "BLOB":
			columnPointer := new([]byte)
			columnPointers = append(columnPointers, columnPointer)
		case "BOOLEAN":
			columnPointer := new(bool)
			columnPointers = append(columnPointers, columnPointer)
		case "JSON":
			if field.Nullable {
				columnPointer := new(*string)
				columnPointers = append(columnPointers, columnPointer)
			} else {
				columnPointer := new(string)
				columnPointers = append(columnPointers, columnPointer)
			}
		case "DATETIME", "TIMESTAMP", "DATE", "TIME", "UUID", "CLOB", "XML":
			if field.Nullable {
				columnPointer := new(*string)
				columnPointers = append(columnPointers, columnPointer)
			} else {
				columnPointer := new(string)
				columnPointers = append(columnPointers, columnPointer)
			}
		default:
			// Default to string for unknown types
			columnPointer := new(string)
			columnPointers = append(columnPointers, columnPointer)
		}
	}

	return &FieldScanner{
		ColumnPointers: columnPointers,
		FieldTypes:     fieldTypes,
		FieldNames:     fieldNames,
		Fields:         tbl.Fields,
	}
}

// ScanToObject converts scanned database values to an object.Object
func (fs *FieldScanner) ScanToObject(tableName string) *object.Object {
	obj := &object.Object{
		TableName: tableName,
		Fields:    make(map[string]any),
	}

	for i, fieldName := range fs.FieldNames {
		field := fs.Fields[fieldName]
		switch strings.ToUpper(fs.FieldTypes[i]) {
		case "TEXT", "VARCHAR", "CHAR":
			if field.Nullable {
				if ptr, ok := fs.ColumnPointers[i].(**string); ok && ptr != nil && *ptr != nil {
					obj.Fields[field.Name] = **ptr
				} else {
					obj.Fields[field.Name] = nil
				}
			} else {
				obj.Fields[field.Name] = *fs.ColumnPointers[i].(*string)
			}
		case "INTEGER", "INT":
			if field.Nullable {
				if ptr, ok := fs.ColumnPointers[i].(**int64); ok && ptr != nil && *ptr != nil {
					obj.Fields[field.Name] = **ptr
				} else {
					obj.Fields[field.Name] = nil
				}
			} else {
				obj.Fields[field.Name] = *fs.ColumnPointers[i].(*int64)
			}
		case "REAL", "FLOAT":
			obj.Fields[field.Name] = *fs.ColumnPointers[i].(*float64)
		case "BLOB":
			obj.Fields[field.Name] = *fs.ColumnPointers[i].(*[]byte)
		case "BOOLEAN":
			obj.Fields[field.Name] = *fs.ColumnPointers[i].(*bool)
		case "JSON":
			if field.Nullable {
				if ptr, ok := fs.ColumnPointers[i].(**string); ok && ptr != nil && *ptr != nil {
					obj.Fields[field.Name] = **ptr
				} else {
					obj.Fields[field.Name] = nil
				}
			} else {
				obj.Fields[field.Name] = *fs.ColumnPointers[i].(*string)
			}
		case "DATETIME", "TIMESTAMP", "DATE", "TIME", "UUID", "CLOB", "XML":
			if field.Nullable {
				if ptr, ok := fs.ColumnPointers[i].(**string); ok && ptr != nil && *ptr != nil {
					obj.Fields[field.Name] = **ptr
				} else {
					obj.Fields[field.Name] = nil
				}
			} else {
				obj.Fields[field.Name] = *fs.ColumnPointers[i].(*string)
			}
		default:
			obj.Fields[field.Name] = *fs.ColumnPointers[i].(*string)
		}
	}

	// Set the ID field from the object fields
	if idValue, exists := obj.Fields["id"]; exists && idValue != nil {
		obj.ID = idValue.(string)
	}

	return obj
}

// GetColumns returns column names for the field scanner
func (fs *FieldScanner) GetColumns() []string {
	columns := make([]string, 0, len(fs.FieldNames))
	for _, fieldName := range fs.FieldNames {
		field := fs.Fields[fieldName]
		columns = append(columns, field.Name)
	}
	return columns
}