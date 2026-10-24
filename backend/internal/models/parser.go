package models

// ColumnReference describes a column mentioned in a SQL query.
type ColumnReference struct {
	Name      string
	TableName string
	Alias     string
	Position  int
}

// TableReference describes a table mentioned in a SQL query.
type TableReference struct {
	Name   string
	Schema string
	Alias  string
}

// ParsedQuery is the structured output of the parser.
type ParsedQuery struct {
	OriginalQuery string
	Columns       []ColumnReference
	Tables        []TableReference
}

// GetColumnNames returns the names of all parsed columns.
func (pq *ParsedQuery) GetColumnNames() []string {
	names := []string{}
	for _, c := range pq.Columns {
		if c.Name != "*" {
			names = append(names, c.Name)
		}
	}
	return names
}

// GetTableNames returns the names of all parsed tables.
func (pq *ParsedQuery) GetTableNames() []string {
	names := []string{}
	for _, t := range pq.Tables {
		names = append(names, t.Name)
	}
	return names
}

// HasColumn reports whether the parsed query references the given column.
func (pq *ParsedQuery) HasColumn(name string) bool {
	for _, c := range pq.Columns {
		if c.Name == name {
			return true
		}
	}
	return false
}
