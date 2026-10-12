// Interactive form: submit a SQL query, see it rewritten.

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
import { Textarea } from "@/components/ui/textarea";
import { useRewrite } from "@/lib/hooks";

const DEFAULT_QUERY = "SELECT tconst, primary_title FROM title_basics LIMIT 10";

export function QueryRewriter() {
  const [query, setQuery] = useState(DEFAULT_QUERY);
  const { call: rewrite, loading, data, error } = useRewrite();

  function handleRewrite() {
    rewrite(query);
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Query Rewriter</CardTitle>
        <CardDescription>
          Old queries are automatically rewritten to use new column names.
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-4">
        <Textarea
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="font-mono text-sm min-h-32"
          placeholder="SELECT primary_title FROM title_basics"
        />

        <Button onClick={handleRewrite} disabled={loading}>
          {loading ? "Rewriting..." : "Rewrite Query"}
        </Button>

        {error && (
          <div className="rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {data && (
          <div className="space-y-2">
            <div>
              <p className="text-xs uppercase tracking-wide text-muted-foreground mb-1">
                Original
              </p>
              <pre className="bg-muted p-3 rounded-md font-mono text-sm overflow-auto">
                {data.original}
              </pre>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-muted-foreground mb-1">
                Rewritten
              </p>
              <pre className="bg-muted p-3 rounded-md font-mono text-sm overflow-auto">
                {data.rewritten}
              </pre>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
