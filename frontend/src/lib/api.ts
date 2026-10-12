// SchemaFlow API Client
import axios, { AxiosRequestConfig } from "axios";
import type {
  RewriteResponse,
  ImpactResponse,
  ImpactRequest,
  Migration,
  LineageRecord,
  RewriteRequest,
} from "./types";

// Axios instance — single config for the whole app
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const client = axios.create({
  baseURL: API_BASE_URL,
  headers: { "Content-Type": "application/json" },
  timeout: 10000,
});

// Supported HTTP methods
type HttpMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

// Generic request function
async function request<T>(
  method: HttpMethod,
  path: string,
  body?: unknown,
): Promise<T> {
  try {
    const config: AxiosRequestConfig = { method, url: path };
    if (body !== undefined) {
      config.data = body;
    }

    const response = await client.request<T>(config);
    return response.data;
  } catch (err) {
    throw new Error(getErrorMessage(err));
  }
}

// Rewrites an old SQL query to use new column names.
export const rewriteQuery = (query: string) =>
  request<RewriteResponse>('POST', '/rewrite', { query } as RewriteRequest);

// Detects views and functions that depend on a column.
export const checkImpact = (data: ImpactRequest) =>
  request<ImpactResponse>("POST", "/impact", data);

// Records a migration event (rename, add, drop).
export const recordMigration = (data: Record<string, unknown>) =>
  request<{ status: string; message: string }>("POST", "/migrate", data);

export const listMigrations = () => request<Migration[]>("GET", "/migrations");

export const listLineage = () => request<LineageRecord[]>("GET", "/lineage");

// Handles network errors, HTTP errors, and unexpected errors.
export function getErrorMessage(err: unknown): string {
  if (axios.isAxiosError(err)) {
    if (err.code === "ECONNREFUSED") {
      return "Backend is not running. Start it with: cd backend && go run .";
    }

    if (err.code === "ECONNABORTED") {
      return "Request timed out after 10 seconds.";
    }

    if (err.response) {
      const data = err.response.data;
      if (typeof data === "string") return data;
      if (data?.message) return String(data.message);
      return `Server error: ${err.response.status} ${err.response.statusText}`;
    }

    return err.message;
  }

  if (err instanceof Error) return err.message;
  return String(err);
}

export { client as apiClient };
