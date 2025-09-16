package schema

type Schema struct {
	DatabaseName string
	Tables       map[string]*TableSchema // Maps table names to their schemas
	FieldOrder   []string                // order to use for iterating Fields
}

type TableSchema struct {
	TableName           string
	Fields              map[string]ColumnData // Maps field names to their definitions
	FieldOrder          []string              // Order of fields for iteration
	PrimaryKey          string                // Name of the primary key field
	Indexes             []string              // List of index names for this table
	UniqueKeys          []string              // List of unique key names for this table
	AutoIncrementFields []string              // List of fields that are auto-incremented
	Comment             string                // Optional comment for the table
}

type ColumnData struct {
	Name          string
	DataType      string
	Nullable      bool
	Default       any // Default value for the column, can be nil
	Comment       string
	Unique        bool
	Index         bool
	AutoIncrement bool
	PrimaryKey    bool // Indicates if this column is a primary key
}

func New() *Schema {
	return &Schema{
		Tables: make(map[string]*TableSchema),
	}
}

func (s *Schema) SetDatabaseName(name string) {
	s.DatabaseName = name
}

func (s *Schema) AddTable(table *TableSchema) {
	if table == nil {
		return
	}
	s.Tables[table.TableName] = table
}

func (s *Schema) GetTable(name string) (*TableSchema, bool) {
	table, exists := s.Tables[name]
	return table, exists
}

func NewTableSchema(name string) *TableSchema {
	return &TableSchema{
		TableName: name,
		Fields:    make(map[string]ColumnData),
	}
}

func (ts *TableSchema) AddField(field ColumnData) {
	if field.Name == "" {
		return
	}
	ts.Fields[field.Name] = field
	ts.FieldOrder = append(ts.FieldOrder, field.Name)
	if field.PrimaryKey {
		ts.PrimaryKey = field.Name
	}
	if field.AutoIncrement {
		ts.AutoIncrementFields = append(ts.AutoIncrementFields, field.Name)
	}
}
