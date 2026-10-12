// Root layout — wraps every page.
// Renders the Header once, then the page content.

import type { Metadata } from 'next';
import { Geist } from 'next/font/google';
import './globals.css';
import { Header } from '@/components/layouts/header';

const geist = Geist({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'SchemaFlow',
  description: 'Zero-downtime PostgreSQL schema evolution with automatic query rewriting',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={`${geist.className} min-h-screen bg-background text-foreground`}>
        <Header />
        <main>{children}</main>
      </body>
    </html>
  );
}