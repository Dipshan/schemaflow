import type { Metadata } from "next";
import { ImpactChecker } from "@/components/impact/ImpactChecker";

export const metadata: Metadata = {
  title: "Impact · SchemaFlow",
  description:
    "Check which views and functions depend on a column before migrating.",
};

export default function ImpactPage() {
  return (
    <div className="max-w-4xl mx-auto px-6 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">
          Impact Checker
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          Check which views, functions, and triggers depend on a column before
          renaming it.
        </p>
      </div>

      <ImpactChecker />
    </div>
  );
}
