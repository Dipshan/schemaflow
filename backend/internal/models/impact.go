// Package models defines shared domain types.
package models

// ColumnDependency describes a single view or function that depends
// on a column.
type ColumnDependency struct {
	ObjectType     string `json:"object_type"`
	SchemaName     string `json:"schema_name"`
	ObjectName     string `json:"object_name"`
	Definition     string `json:"definition"`
	DependencyType string `json:"dependency_type"`
}

// ImpactReport is the response returned by the /impact endpoint.
type ImpactReport struct {
	Schema       string             `json:"schema"`
	Table        string             `json:"table"`
	Column       string             `json:"column"`
	Dependencies []ColumnDependency `json:"dependencies"`
	Total        int                `json:"total"`
}

// HasDependencies reports whether the report contains any dependencies.
func (r *ImpactReport) HasDependencies() bool {
	return r.Total > 0
}
