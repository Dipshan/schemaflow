// Rewrite API
export interface RewriteRequest {
  query: string;
}

export interface RewriteResponse {
  original: string;
  rewritten: string;
}

// Impact API
export interface ImpactRequest {
  schema: string;
  table: string;
  column: string;
}

export interface Dependency {
  ObjectType: string;
  SchemaName: string;
  ObjectName: string;
  Definition: string;
  DependencyType: string;
}

export interface ImpactResponse {
  column: string;
  table: string;
  schema: string;
  dependencies: Dependency[];
  total: number;
  has_dependencies: boolean;
  warning?: string;
  message?: string;
}

// Migrations API (for future pages)
export interface Migration {
  event_id: string;
  version: number;
  event_type: string;
  table_name: string;
  old_name: string;
  new_name: string;
  applied_at: string;
  applied_by: string;
}

// Lineage API (for future pages)
export interface LineageRecord {
  lineage_id: number;
  schema_event_id: string;
  schema_name: string;
  table_name: string;
  old_column_name: string;
  new_column_name: string;
  is_current: boolean;
  valid_from: string;
  valid_to: string | null;
}
