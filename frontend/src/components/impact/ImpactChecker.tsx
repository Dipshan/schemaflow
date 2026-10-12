// Interactive form: check what depends on a column before migrating.

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useImpact } from "@/lib/hooks";
import type { ImpactResponse } from "@/lib/types";

const DEFAULT_SCHEMA = "public";
const DEFAULT_TABLE = "title_basics";
const DEFAULT_COLUMN = "movie_title";

export function ImpactChecker() {
  const [schema, setSchema] = useState(DEFAULT_SCHEMA);
  const [table, setTable] = useState(DEFAULT_TABLE);
  const [column, setColumn] = useState(DEFAULT_COLUMN);

  const { call: checkImpact, loading, data, error } = useImpact();

  function handleCheck() {
    checkImpact({ schema, table, column });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Impact Checker</CardTitle>
        <CardDescription>
          See which views, functions, and triggers depend on a column before
          renaming it.
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          <div>
            <label className="text-xs uppercase tracking-wide text-muted-foreground">
              Schema
            </label>
            <Input
              value={schema}
              onChange={(e) => setSchema(e.target.value)}
              className="mt-1 font-mono text-sm"
            />
          </div>
          <div>
            <label className="text-xs uppercase tracking-wide text-muted-foreground">
              Table
            </label>
            <Input
              value={table}
              onChange={(e) => setTable(e.target.value)}
              className="mt-1 font-mono text-sm"
            />
          </div>
          <div>
            <label className="text-xs uppercase tracking-wide text-muted-foreground">
              Column
            </label>
            <Input
              value={column}
              onChange={(e) => setColumn(e.target.value)}
              className="mt-1 font-mono text-sm"
            />
          </div>
        </div>

        <Button
          onClick={handleCheck}
          disabled={loading}
          className={loading ? "cursor-not-allowed" : "cursor-pointer"}
        >
          {loading ? "Checking..." : "Check Impact"}
        </Button>

        {error && (
          <div className="rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {data && <ImpactResult data={data} />}
      </CardContent>
    </Card>
  );
}

function ImpactResult({ data }: { data: ImpactResponse }) {
  if (!data.has_dependencies) {
    return (
      <div className="rounded-md border border-green-300 bg-green-50 p-3 text-sm text-green-800">
        No dependencies found. Safe to rename this column.
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="rounded-md border border-yellow-300 bg-yellow-50 p-3 text-sm text-yellow-900">
        {data.total} {data.total === 1 ? "dependency" : "dependencies"} found.
        Renaming this column may break them.
      </div>

      <div className="border rounded-md divide-y">
        {data.dependencies.map((dep, i) => (
          <div key={i} className="p-3">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                {dep.ObjectType}
              </span>
              <span className="font-mono text-sm">
                {dep.SchemaName}.{dep.ObjectName}
              </span>
            </div>
            <pre className="bg-muted p-2 rounded text-xs font-mono overflow-auto max-h-32">
              {dep.Definition}
            </pre>
          </div>
        ))}
      </div>
    </div>
  );
}
