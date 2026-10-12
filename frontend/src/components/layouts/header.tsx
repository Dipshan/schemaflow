// Top navigation bar for all pages.
// Server Component — no state, no events, no hooks.

import Link from "next/link";

const navLinks = [
  { href: "/rewrite", label: "Rewrite" },
  { href: "/impact", label: "Impact" },
  { href: "/migrations", label: "Migrations" },
  { href: "/lineage", label: "Lineage" },
];

export function Header() {
  return (
    <header className="border-b">
      <div className="max-w-6xl mx-auto p-6 flex items-center justify-between">
        <Link href="/" className="text-lg font-semibold hover:opacity-80">
          SchemaFlow
        </Link>

        <nav className="flex gap-6 text-sm text-muted-foreground">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="hover:text-foreground transition-colors"
            >
              {link.label}
            </Link>
          ))}
        </nav>
      </div>
    </header>
  );
}
