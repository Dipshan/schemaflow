// Homepage route.
// Redirects to /rewrite — the main entry point of SchemaFlow.

import { redirect } from "next/navigation";

export default function Home() {
  redirect("/rewrite");
}
