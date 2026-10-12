# SchemaFlow

**Author:** Deepshan Adhikari (deepadh@siue.edu)   
**Institution:** Southern Illinois University Edwardsville   
**Course:** Master of Science in Computer Science - Master's Project  

A zero-downtime PostgreSQL schema evolution tool with automatic query rewriting.

## What It Is

SchemaFlow allows developers to rename database columns without breaking existing application queries. It records every migration, tracks column lineage across renames, rewrites old SQL queries to use new column names, and predicts which views, functions, and triggers depend on a column before a migration is applied.

The problem is real and well documented. In August 2025, GitHub experienced a 4 percent API error rate because a dropped column was still referenced by their ORM layer. Traditional migration tools handle the schema change itself but leave applications to update their queries manually. SchemaFlow bridges this gap.

## How It Works

1. **Record** — A developer sends a rename request to `/migrate`. SchemaFlow stores the event in `schema_events` (audit log) and the old-to-new mapping in `column_lineage`.
2. **Parse** — When an old query arrives, `/parse` extracts the columns and tables it references.
3. **Rewrite** — `/rewrite` looks up each column in `column_lineage` and replaces old names with current names, so the query keeps working.
4. **Predict** — Before migrating, `/impact` queries PostgreSQL's system catalog (`pg_catalog`) to return all views, functions, and triggers that depend on the column.

The workflow is: check impact first, fix dependencies, then migrate, then let SchemaFlow rewrite queries automatically.

## Project Structure

```
schemaflow/
├── backend/           Go + PostgreSQL REST API
│   ├── main.go
│   ├── database.go
│   ├── migration.go
│   ├── rewriter.go
│   ├── impact.go
│   └── README.md
├── frontend/          Next.js dashboard
│   ├── src/
│   │   ├── app/       Four pages (rewrite, impact, migrations, lineage)
│   │   ├── components/
│   │   └── lib/
│   └── README.md
└── docs/
    ├── experiments/   Experiment scripts and results
    └── reports/       Test reports and paper drafts
```

## System Architecture

```
Application (old SQL query)
        |
        v
SchemaFlow Backend (Go, port 8080)
        |
        |  /parse    -> extract columns
        |  /rewrite  -> replace old names
        |  /impact   -> detect dependencies
        |  /migrate  -> record changes
        v
PostgreSQL 16 (schemaflow_dev)
        |
        |  schema_events    (audit log)
        |  column_lineage   (old -> new mappings)
        |  title_basics, name_basics, title_ratings, title_principals
        v
SchemaFlow Frontend (Next.js, port 3000)
```

## Technology Stack

**Backend**
- Go 1.26 with `net/http` and `pgx/v5`
- PostgreSQL 16 with UUID primary keys, foreign keys, and partial indexes

**Frontend**
- Next.js 16 with App Router and Turbopack
- React 19 with React Compiler
- TypeScript
- Tailwind CSS 4
- shadcn/ui components
- Axios for HTTP requests

**Data**
- IMDb dataset
- Four tables: `title_basics`, `name_basics`, `title_ratings`, `title_principals`
- 25,000 rows per table, 100,000 rows total
- Three views and two functions created for impact testing

## API Endpoints

| Method | Path          | Description                                         |
|--------|---------------|-----------------------------------------------------|
| POST   | `/migrate`    | Record a schema change (rename/add/drop)            |
| POST   | `/parse`      | Extract columns and tables from a SQL query         |
| POST   | `/rewrite`    | Rewrite a query to use new column names             |
| POST   | `/impact`     | Detect views and functions that depend on a column  |
| GET    | `/migrations` | List all recorded migrations                        |
| GET    | `/lineage`    | List all column lineage records                     |

## Database Schema

**`schema_events`** — immutable audit log of every migration

| Column       | Type        | Purpose                                            |
|--------------|-------------|----------------------------------------------------|
| `event_id`   | UUID (PK)   | Unique migration identifier                        |
| `version`    | INTEGER     | Sequential version number                          |
| `event_type` | TEXT        | RENAME_COLUMN, ADD_COLUMN, DROP_COLUMN             |
| `table_name` | TEXT        | Affected table                                     |
| `old_name`   | TEXT        | Column name before the change                      |
| `new_name`   | TEXT        | Column name after the change                       |
| `applied_at` | TIMESTAMPTZ | When the migration was recorded                    |
| `applied_by` | TEXT        | Who applied the migration                          |

**`column_lineage`** — old-to-new column name mappings

| Column             | Type           | Purpose                                  |
|--------------------|----------------|------------------------------------------|
| `lineage_id`       | BIGSERIAL (PK) | Unique lineage record identifier         |
| `schema_event_id`  | UUID (FK)      | Links to `schema_events.event_id`        |
| `schema_name`      | TEXT           | Schema of the affected table             |
| `table_name`       | TEXT           | Table of the affected column             |
| `old_column_name`  | TEXT           | Original column name                     |
| `new_column_name`  | TEXT           | Current column name                      |
| `is_current`       | BOOLEAN        | True for the active mapping              |
| `valid_from`       | TIMESTAMPTZ    | When the mapping became active           |
| `valid_to`         | TIMESTAMPTZ    | When it was superseded (NULL if current) |

## Setup

### Prerequisites

- Go 1.26 or higher
- PostgreSQL 16 or higher
- Node.js 20 or higher
- A database named `schemaflow_dev` with the IMDb tables loaded

### Backend

```bash
cd backend
go mod tidy
go run .
```

Server starts on **http://localhost:8080**.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Dashboard runs on **http://localhost:3000**.

## Experiments

The system was evaluated on real IMDb data (100,000 rows across four tables) with the following experiments.

**Experiment 1 — Correctness**
Confirms that a rewritten query returns the same rows as the original query. A query on `title_basics` was run against the original schema, the column `primary_title` was renamed to `movie_title` via `/migrate`, the query was rewritten via `/rewrite`, and the results were compared row by row. The output matched across all rows, with only the column header changing as expected.

**Experiment 2 — Performance**
Measures rewrite latency. Ten thousand calls to `/rewrite` were issued in sequence with each call timed. Average, minimum, maximum, and P99 latencies were recorded to evaluate whether rewriting meets production requirements.

**Experiment 3 — Impact**
Confirms that `/impact` detects dependent objects before a rename. A view (`v_long_movies`) and a function (`fn_get_runtime`) were created on `title_basics`, and `/impact` was called with the target column. The report lists every object that depends on the column. A finding from this experiment is that impact prediction must run before the rename, not after, because PostgreSQL removes dependency records when a column is renamed.

**Experiment 4 — Concurrency**
Confirms that a migration can run while concurrent readers are active. Ten goroutines performed reads from `title_basics` while an `ALTER TABLE ADD COLUMN` was executed mid-run. No deadlocks and no failed reads occurred.

**Experiment 5 — Robustness**
Tests SchemaFlow against the ten perturbation types defined by the EvoSchema benchmark: add column, remove column, rename column, split column, merge column, change type, add table, remove table, rename table, and split table. Results are documented in `docs/experiments/05_robustness/`.

## Related Work

- **PRISM** (IEEE ICDE 2009) — proved query rewriting for schema evolution is feasible using SQL-based schema modification operators.
- **PRIMA** (ACM SIGMOD 2009) — extended PRISM to support historical queries across schema versions.
- **Panta Rhei** (Microsoft Research, 2009) — a complete schema evolution framework for Microsoft SQL Server.
- **ToMAS** (IEEE ICDE 2004) — automatic adaptation of mappings when schemas evolve.
- **EvoSchema** (VLDB 2025) — benchmark of ten schema perturbation types for testing text-to-SQL robustness.
- **pgroll** (2023) — production PostgreSQL migration tool with zero-downtime migrations but no query rewriting.
- **Flyway** — version-based migration tool with downtime and no query rewriting.

SchemaFlow is the first tool that combines zero-downtime migrations with automatic query rewriting for PostgreSQL.