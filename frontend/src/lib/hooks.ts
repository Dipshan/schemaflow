// SchemaFlow React Hooks

"use client";

import { useState, useCallback, useRef } from "react";
import {
  getErrorMessage,
  rewriteQuery,
  checkImpact,
  recordMigration,
  listMigrations,
  listLineage,
} from "./api";

interface UseApiCallResult<TArgs extends unknown[], TResult> {
  call: (...args: TArgs) => Promise<void>;
  loading: boolean;
  data: TResult | null;
  error: string | null;
  reset: () => void;
}

// Generic hook — the engine. Every API hook is built on this.
export function useApiCall<TArgs extends unknown[], TResult>(
  apiFunction: (...args: TArgs) => Promise<TResult>,
): UseApiCallResult<TArgs, TResult> {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<TResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const mountedRef = useRef(true);

  const call = useCallback(
    async (...args: TArgs) => {
      setLoading(true);
      setError(null);

      try {
        const result = await apiFunction(...args);
        if (mountedRef.current) setData(result);
      } catch (err) {
        if (mountedRef.current) setError(getErrorMessage(err));
      } finally {
        if (mountedRef.current) setLoading(false);
      }
    },
    [apiFunction],
  );

  const reset = useCallback(() => {
    setData(null);
    setError(null);
  }, []);

  return { call, loading, data, error, reset };
}

export const useRewrite = () => useApiCall(rewriteQuery);
export const useImpact = () => useApiCall(checkImpact);
export const useMigrate = () => useApiCall(recordMigration);
export const useMigrations = () => useApiCall(listMigrations);
export const useLineage = () => useApiCall(listLineage);
