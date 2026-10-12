// Displays all migrations from schema_events.

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
import { useMigrations } from "@/lib/hooks";

export function MigrationTable() {
  const { call: load, loading, data, error } = useMigrations();

  // Fetch once on mount
  useEffect(() => {
    load();
  }, [load]);

  return (
    <Card>
      <CardContent>
        {loading && (
          <p className="text-sm text-muted-foreground">Loading migrations...</p>
        )}

        {error && (
          <div className="rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {data && data.length === 0 && (
          <p className="text-sm text-muted-foreground">
            No migrations yet. Record one via the /migrate API.
          </p>
        )}

        {data && data.length > 0 && (
          <div className="border rounded-md overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">Ver</TableHead>
                  <TableHead>Event</TableHead>
                  <TableHead>Table</TableHead>
                  <TableHead>Old → New</TableHead>
                  <TableHead className="text-right">By</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.map((m) => (
                  <TableRow key={m.event_id}>
                    <TableCell className="font-mono text-xs">
                      {m.version}
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary" className="text-xs">
                        {m.event_type}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {m.table_name}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {m.old_name} → {m.new_name}
                    </TableCell>
                    <TableCell className="text-right text-xs text-muted-foreground">
                      {m.applied_by}
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
