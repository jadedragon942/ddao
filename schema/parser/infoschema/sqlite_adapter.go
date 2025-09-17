package infoschema

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jadedragon942/ddao/schema"
)

type SQLiteAdapter struct {
	db *sql.DB
}

func NewSQLiteAdapter(db *sql.DB) *SQLiteAdapter {
	return &SQLiteAdapter{db: db}
}

func (s *SQLiteAdapter) ParseSchema(databaseName string) (*schema.Schema, error) {
	sch := schema.New()
	sch.SetDatabaseName(databaseName)

	tables, err := s.getTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	for _, tableName := range tables {
		tableSchema, err := s.parseTable(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse table %s: %w", tableName, err)
		}
		sch.AddTable(tableSchema)
	}

	return sch, nil
}

func (s *SQLiteAdapter) getTables() ([]string, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

func (s *SQLiteAdapter) parseTable(tableName string) (*schema.TableSchema, error) {
	tableSchema := schema.NewTableSchema(tableName)

	columns, err := s.getColumns(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	for _, column := range columns {
		columnData := s.convertColumn(column)
		tableSchema.AddField(columnData)
	}

	indexes, err := s.getIndexes(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}

	s.processIndexes(tableSchema, indexes)

	s.updateColumnConstraints(tableSchema, indexes)

	return tableSchema, nil
}

type SQLiteColumn struct {
	Name         string
	Type         string
	NotNull      bool
	DefaultValue interface{}
	PrimaryKey   bool
}

type SQLiteIndex struct {
	Name    string
	Unique  bool
	Columns []string
}

func (s *SQLiteAdapter) getColumns(tableName string) ([]SQLiteColumn, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []SQLiteColumn
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk bool
		var defaultValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return nil, err
		}

		columns = append(columns, SQLiteColumn{
			Name:         name,
			Type:         dataType,
			NotNull:      notNull,
			DefaultValue: defaultValue,
			PrimaryKey:   pk,
		})
	}

	return columns, rows.Err()
}

func (s *SQLiteAdapter) getIndexes(tableName string) ([]SQLiteIndex, error) {
	query := fmt.Sprintf("PRAGMA index_list(%s)", tableName)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type indexInfo struct {
		name   string
		unique bool
	}

	var indexInfos []indexInfo
	for rows.Next() {
		var seq int
		var name string
		var unique bool
		var origin, partial string

		err := rows.Scan(&seq, &name, &unique, &origin, &partial)
		if err != nil {
			return nil, err
		}

		indexInfos = append(indexInfos, indexInfo{
			name:   name,
			unique: unique,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var indexes []SQLiteIndex
	for _, info := range indexInfos {
		columns, err := s.getIndexColumns(info.name)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, SQLiteIndex{
			Name:    info.name,
			Unique:  info.unique,
			Columns: columns,
		})
	}

	return indexes, nil
}

func (s *SQLiteAdapter) getIndexColumns(indexName string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA index_info(%s)", indexName)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var seqno, cid int
		var name string

		err := rows.Scan(&seqno, &cid, &name)
		if err != nil {
			return nil, err
		}

		columns = append(columns, name)
	}

	return columns, rows.Err()
}

func (s *SQLiteAdapter) convertColumn(column SQLiteColumn) schema.ColumnData {
	columnData := schema.ColumnData{
		Name:       column.Name,
		DataType:   column.Type,
		Nullable:   !column.NotNull,
		PrimaryKey: column.PrimaryKey,
	}

	if column.DefaultValue != nil {
		columnData.Default = column.DefaultValue
	}

	columnData.AutoIncrement = column.PrimaryKey && strings.ToUpper(column.Type) == "INTEGER"

	return columnData
}

func (s *SQLiteAdapter) processIndexes(tableSchema *schema.TableSchema, indexes []SQLiteIndex) {
	for _, index := range indexes {
		if strings.HasPrefix(index.Name, "sqlite_autoindex_") {
			continue
		}

		tableSchema.Indexes = append(tableSchema.Indexes, index.Name)
		if index.Unique {
			tableSchema.UniqueKeys = append(tableSchema.UniqueKeys, index.Name)
		}
	}
}

func (s *SQLiteAdapter) ParseTableFromName(databaseName, tableName string) (*schema.TableSchema, error) {
	tables, err := s.getTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	for _, table := range tables {
		if table == tableName {
			return s.parseTable(tableName)
		}
	}

	return nil, fmt.Errorf("table %s not found in database %s", tableName, databaseName)
}

func (s *SQLiteAdapter) GetTableNames(databaseName string) ([]string, error) {
	return s.getTables()
}

func (s *SQLiteAdapter) updateColumnConstraints(tableSchema *schema.TableSchema, indexes []SQLiteIndex) {
	for _, index := range indexes {
		if index.Unique && len(index.Columns) == 1 {
			columnName := index.Columns[0]
			if field, exists := tableSchema.Fields[columnName]; exists {
				field.Unique = true
				field.Index = true
				tableSchema.Fields[columnName] = field
			}
		}
		if len(index.Columns) >= 1 {
			for _, columnName := range index.Columns {
				if field, exists := tableSchema.Fields[columnName]; exists {
					field.Index = true
					tableSchema.Fields[columnName] = field
				}
			}
		}
	}
}
