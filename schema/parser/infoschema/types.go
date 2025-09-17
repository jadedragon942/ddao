package infoschema

import "time"

type InfoSchemaTable struct {
	TableCatalog   string     `db:"table_catalog"`
	TableSchema    string     `db:"table_schema"`
	TableName      string     `db:"table_name"`
	TableType      string     `db:"table_type"`
	Engine         string     `db:"engine"`
	Version        *int64     `db:"version"`
	RowFormat      string     `db:"row_format"`
	TableRows      *int64     `db:"table_rows"`
	AvgRowLength   *int64     `db:"avg_row_length"`
	DataLength     *int64     `db:"data_length"`
	MaxDataLength  *int64     `db:"max_data_length"`
	IndexLength    *int64     `db:"index_length"`
	DataFree       *int64     `db:"data_free"`
	AutoIncrement  *int64     `db:"auto_increment"`
	CreateTime     *time.Time `db:"create_time"`
	UpdateTime     *time.Time `db:"update_time"`
	CheckTime      *time.Time `db:"check_time"`
	TableCollation string     `db:"table_collation"`
	Checksum       *int64     `db:"checksum"`
	CreateOptions  string     `db:"create_options"`
	TableComment   string     `db:"table_comment"`
}

type InfoSchemaColumn struct {
	TableCatalog           string  `db:"table_catalog"`
	TableSchema            string  `db:"table_schema"`
	TableName              string  `db:"table_name"`
	ColumnName             string  `db:"column_name"`
	OrdinalPosition        int     `db:"ordinal_position"`
	ColumnDefault          *string `db:"column_default"`
	IsNullable             string  `db:"is_nullable"`
	DataType               string  `db:"data_type"`
	CharacterMaximumLength *int    `db:"character_maximum_length"`
	CharacterOctetLength   *int    `db:"character_octet_length"`
	NumericPrecision       *int    `db:"numeric_precision"`
	NumericScale           *int    `db:"numeric_scale"`
	DatetimePrecision      *int    `db:"datetime_precision"`
	CharacterSetName       *string `db:"character_set_name"`
	CollationName          *string `db:"collation_name"`
	ColumnType             string  `db:"column_type"`
	ColumnKey              string  `db:"column_key"`
	Extra                  string  `db:"extra"`
	Privileges             string  `db:"privileges"`
	ColumnComment          string  `db:"column_comment"`
	GenerationExpression   string  `db:"generation_expression"`
	SrsId                  *int    `db:"srs_id"`
}

type InfoSchemaKeyColumn struct {
	ConstraintCatalog          string  `db:"constraint_catalog"`
	ConstraintSchema           string  `db:"constraint_schema"`
	ConstraintName             string  `db:"constraint_name"`
	TableCatalog               string  `db:"table_catalog"`
	TableSchema                string  `db:"table_schema"`
	TableName                  string  `db:"table_name"`
	ColumnName                 string  `db:"column_name"`
	OrdinalPosition            int     `db:"ordinal_position"`
	PositionInUniqueConstraint *int    `db:"position_in_unique_constraint"`
	ReferencedTableSchema      *string `db:"referenced_table_schema"`
	ReferencedTableName        *string `db:"referenced_table_name"`
	ReferencedColumnName       *string `db:"referenced_column_name"`
}

type InfoSchemaIndex struct {
	TableCatalog string  `db:"table_catalog"`
	TableSchema  string  `db:"table_schema"`
	TableName    string  `db:"table_name"`
	NonUnique    int     `db:"non_unique"`
	IndexSchema  string  `db:"index_schema"`
	IndexName    string  `db:"index_name"`
	SeqInIndex   int     `db:"seq_in_index"`
	ColumnName   string  `db:"column_name"`
	Collation    string  `db:"collation"`
	Cardinality  *int64  `db:"cardinality"`
	SubPart      *int    `db:"sub_part"`
	Packed       *string `db:"packed"`
	Nullable     string  `db:"nullable"`
	IndexType    string  `db:"index_type"`
	Comment      string  `db:"comment"`
	IndexComment string  `db:"index_comment"`
	IsVisible    string  `db:"is_visible"`
	Expression   *string `db:"expression"`
}

type InfoSchemaConstraint struct {
	ConstraintCatalog string `db:"constraint_catalog"`
	ConstraintSchema  string `db:"constraint_schema"`
	ConstraintName    string `db:"constraint_name"`
	TableCatalog      string `db:"table_catalog"`
	TableSchema       string `db:"table_schema"`
	TableName         string `db:"table_name"`
	ConstraintType    string `db:"constraint_type"`
	Enforced          string `db:"enforced"`
}

type InfoSchemaReferentialConstraint struct {
	ConstraintCatalog       string `db:"constraint_catalog"`
	ConstraintSchema        string `db:"constraint_schema"`
	ConstraintName          string `db:"constraint_name"`
	UniqueConstraintCatalog string `db:"unique_constraint_catalog"`
	UniqueConstraintSchema  string `db:"unique_constraint_schema"`
	UniqueConstraintName    string `db:"unique_constraint_name"`
	MatchOption             string `db:"match_option"`
	UpdateRule              string `db:"update_rule"`
	DeleteRule              string `db:"delete_rule"`
	TableName               string `db:"table_name"`
	ReferencedTableName     string `db:"referenced_table_name"`
}
