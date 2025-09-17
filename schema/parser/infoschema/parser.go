package infoschema

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jadedragon942/ddao/schema"
)

type Parser struct {
	db *sql.DB
}

func NewParser(db *sql.DB) *Parser {
	return &Parser{db: db}
}

func (p *Parser) ParseSchema(databaseName string) (*schema.Schema, error) {
	s := schema.New()
	s.SetDatabaseName(databaseName)

	tables, err := p.getTables(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	for _, table := range tables {
		tableSchema, err := p.parseTable(table)
		if err != nil {
			return nil, fmt.Errorf("failed to parse table %s: %w", table.TableName, err)
		}
		s.AddTable(tableSchema)
	}

	return s, nil
}

func (p *Parser) getTables(databaseName string) ([]InfoSchemaTable, error) {
	query := `
		SELECT
			table_catalog, table_schema, table_name, table_type,
			COALESCE(engine, '') as engine,
			version,
			COALESCE(row_format, '') as row_format,
			table_rows, avg_row_length, data_length, max_data_length,
			index_length, data_free, auto_increment,
			create_time, update_time, check_time,
			COALESCE(table_collation, '') as table_collation,
			checksum,
			COALESCE(create_options, '') as create_options,
			COALESCE(table_comment, '') as table_comment
		FROM information_schema.tables
		WHERE table_schema = ? AND table_type = 'BASE TABLE'
		ORDER BY table_name`

	rows, err := p.db.Query(query, databaseName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []InfoSchemaTable
	for rows.Next() {
		var table InfoSchemaTable
		err := rows.Scan(
			&table.TableCatalog, &table.TableSchema, &table.TableName, &table.TableType,
			&table.Engine, &table.Version, &table.RowFormat,
			&table.TableRows, &table.AvgRowLength, &table.DataLength, &table.MaxDataLength,
			&table.IndexLength, &table.DataFree, &table.AutoIncrement,
			&table.CreateTime, &table.UpdateTime, &table.CheckTime,
			&table.TableCollation, &table.Checksum, &table.CreateOptions, &table.TableComment,
		)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

func (p *Parser) parseTable(table InfoSchemaTable) (*schema.TableSchema, error) {
	tableSchema := schema.NewTableSchema(table.TableName)
	tableSchema.Comment = table.TableComment

	columns, err := p.getColumns(table.TableSchema, table.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	for _, column := range columns {
		columnData := p.convertColumn(column)
		tableSchema.AddField(columnData)
	}

	indexes, err := p.getIndexes(table.TableSchema, table.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}

	p.processIndexes(tableSchema, indexes)

	return tableSchema, nil
}

func (p *Parser) getColumns(schemaName, tableName string) ([]InfoSchemaColumn, error) {
	query := `
		SELECT
			table_catalog, table_schema, table_name, column_name,
			ordinal_position, column_default, is_nullable, data_type,
			character_maximum_length, character_octet_length,
			numeric_precision, numeric_scale, datetime_precision,
			character_set_name, collation_name, column_type,
			column_key, extra, privileges, column_comment,
			COALESCE(generation_expression, '') as generation_expression,
			srs_id
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position`

	rows, err := p.db.Query(query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []InfoSchemaColumn
	for rows.Next() {
		var column InfoSchemaColumn
		err := rows.Scan(
			&column.TableCatalog, &column.TableSchema, &column.TableName, &column.ColumnName,
			&column.OrdinalPosition, &column.ColumnDefault, &column.IsNullable, &column.DataType,
			&column.CharacterMaximumLength, &column.CharacterOctetLength,
			&column.NumericPrecision, &column.NumericScale, &column.DatetimePrecision,
			&column.CharacterSetName, &column.CollationName, &column.ColumnType,
			&column.ColumnKey, &column.Extra, &column.Privileges, &column.ColumnComment,
			&column.GenerationExpression, &column.SrsId,
		)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns, rows.Err()
}

func (p *Parser) getIndexes(schemaName, tableName string) ([]InfoSchemaIndex, error) {
	query := `
		SELECT
			table_catalog, table_schema, table_name, non_unique,
			index_schema, index_name, seq_in_index, column_name,
			COALESCE(collation, '') as collation, cardinality, sub_part,
			packed, nullable, index_type,
			COALESCE(comment, '') as comment,
			COALESCE(index_comment, '') as index_comment,
			COALESCE(is_visible, 'YES') as is_visible,
			expression
		FROM information_schema.statistics
		WHERE table_schema = ? AND table_name = ?
		ORDER BY index_name, seq_in_index`

	rows, err := p.db.Query(query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []InfoSchemaIndex
	for rows.Next() {
		var index InfoSchemaIndex
		err := rows.Scan(
			&index.TableCatalog, &index.TableSchema, &index.TableName, &index.NonUnique,
			&index.IndexSchema, &index.IndexName, &index.SeqInIndex, &index.ColumnName,
			&index.Collation, &index.Cardinality, &index.SubPart,
			&index.Packed, &index.Nullable, &index.IndexType,
			&index.Comment, &index.IndexComment, &index.IsVisible, &index.Expression,
		)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, index)
	}

	return indexes, rows.Err()
}

func (p *Parser) convertColumn(column InfoSchemaColumn) schema.ColumnData {
	columnData := schema.ColumnData{
		Name:     column.ColumnName,
		DataType: column.DataType,
		Nullable: column.IsNullable == "YES",
		Comment:  column.ColumnComment,
	}

	if column.ColumnDefault != nil {
		columnData.Default = *column.ColumnDefault
	}

	columnData.PrimaryKey = column.ColumnKey == "PRI"
	columnData.Unique = column.ColumnKey == "UNI"
	columnData.Index = column.ColumnKey == "MUL" || column.ColumnKey == "UNI" || column.ColumnKey == "PRI"

	columnData.AutoIncrement = strings.Contains(strings.ToUpper(column.Extra), "AUTO_INCREMENT")

	return columnData
}

func (p *Parser) processIndexes(tableSchema *schema.TableSchema, indexes []InfoSchemaIndex) {
	indexMap := make(map[string][]string)
	uniqueIndexes := make(map[string]bool)

	for _, index := range indexes {
		if index.IndexName == "PRIMARY" {
			continue
		}

		indexMap[index.IndexName] = append(indexMap[index.IndexName], index.ColumnName)
		if index.NonUnique == 0 {
			uniqueIndexes[index.IndexName] = true
		}
	}

	for indexName := range indexMap {
		tableSchema.Indexes = append(tableSchema.Indexes, indexName)
		if uniqueIndexes[indexName] {
			tableSchema.UniqueKeys = append(tableSchema.UniqueKeys, indexName)
		}
	}
}

func (p *Parser) ParseTableFromName(databaseName, tableName string) (*schema.TableSchema, error) {
	tables, err := p.getTables(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	for _, table := range tables {
		if table.TableName == tableName {
			return p.parseTable(table)
		}
	}

	return nil, fmt.Errorf("table %s not found in database %s", tableName, databaseName)
}

func (p *Parser) GetDatabaseNames() ([]string, error) {
	query := "SELECT DISTINCT table_schema FROM information_schema.tables WHERE table_type = 'BASE TABLE' ORDER BY table_schema"

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}

	return databases, rows.Err()
}

func (p *Parser) GetTableNames(databaseName string) ([]string, error) {
	query := "SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE' ORDER BY table_name"

	rows, err := p.db.Query(query, databaseName)
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
