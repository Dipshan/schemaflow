// /migrations route.
// Server Component — renders the MigrationTable client component.

import type { Metadata } from "next";
import { MigrationTable } from "@/components/migrations/MigrationTable";

export const metadata: Metadata = {
  title: "Migrations · SchemaFlow",
  description: "Every schema change recorded by SchemaFlow.",
};

export default function MigrationsPage() {
  return (
    <div className="max-w-6xl mx-auto px-6 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">Migrations</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Every schema change recorded by SchemaFlow, most recent first.
        </p>
      </div>

      <MigrationTable />
    </div>
  );
}
