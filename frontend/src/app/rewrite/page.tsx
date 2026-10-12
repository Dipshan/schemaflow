import type { Metadata } from "next";
import { QueryRewriter } from "@/components/rewrite/QueryRewriter";

export const metadata: Metadata = {
  title: "Rewrite | SchemaFlow",
  description: "Rewrite old SQL queries to use new column names.",
};

export default function RewritePage() {
  return (
    <div className="max-w-4xl mx-auto px-6 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">Rewrite Query</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Paste a query that uses an old column name. SchemaFlow will rewrite it
          to use the new name.
        </p>
      </div>

      <QueryRewriter />
    </div>
  );
}
