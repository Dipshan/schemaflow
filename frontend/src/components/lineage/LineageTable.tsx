// Displays all column lineage records from column_lineage.
// Client Component — fetches data on mount via the useLineage hook.

"use client";

import { useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useLineage } from "@/lib/hooks";

export function LineageTable() {
  const { call: load, loading, data, error } = useLineage();

  useEffect(() => {
    load();
  }, [load]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Column Lineage</CardTitle>
        <CardDescription>
          Every column rename tracked by SchemaFlow. Current mappings are
          marked.
        </CardDescription>
      </CardHeader>

      <CardContent>
        {loading && (
          <p className="text-sm text-muted-foreground">Loading lineage...</p>
        )}

        {error && (
          <div className="rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {data && data.length === 0 && (
          <p className="text-sm text-muted-foreground">
            No lineage records yet. Rename a column via /migrate to create one.
          </p>
        )}

        {data && data.length > 0 && (
          <div className="border rounded-md overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">ID</TableHead>
                  <TableHead>Table</TableHead>
                  <TableHead>Old Column</TableHead>
                  <TableHead>New Column</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">From</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.map((record) => (
                  <TableRow key={record.lineage_id}>
                    <TableCell className="font-mono text-xs">
                      {record.lineage_id}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {record.table_name}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {record.old_column_name}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {record.new_column_name}
                    </TableCell>
                    <TableCell>
                      {record.is_current ? (
                        <Badge className="text-xs">Current</Badge>
                      ) : (
                        <Badge variant="secondary" className="text-xs">
                          Superseded
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right text-xs text-muted-foreground">
                      {formatDate(record.valid_from)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function formatDate(iso: string): string {
  if (!iso) return "—";
  try {
    const d = new Date(iso);
    return d.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}
