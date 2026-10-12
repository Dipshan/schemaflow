// /lineage route.
// Server Component — renders the LineageTable client component.

import { LineageTable } from "@/components/lineage/LineageTable";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Lineage · SchemaFlow",
  description: "Every column rename tracked by SchemaFlow.",
};

export default function LineagePage() {
  return (
    <div className="max-w-6xl mx-auto px-6 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">
          Column Lineage
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          Every column rename tracked by SchemaFlow. Current mappings are
          marked.
        </p>
      </div>

      <LineageTable />
    </div>
  );
}
