"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";

export function Providers({ children }: { children: React.ReactNode }) {
  // networkMode "always": we talk to a same-LAN backend, so the browser's
  // online/offline signal is meaningless — without this, one failed fetch
  // during page load leaves queries paused ("loading" forever) instead of
  // retrying.
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: { networkMode: "always", retry: 3 },
          mutations: { networkMode: "always" },
        },
      }),
  );
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}
