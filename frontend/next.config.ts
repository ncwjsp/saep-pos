import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // A stray lockfile in the user home dir makes workspace-root inference
  // pick the wrong directory; pin it here.
  turbopack: {
    root: __dirname,
  },
};

export default nextConfig;
