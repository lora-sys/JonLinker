import type { NextConfig } from "next";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

const nextConfig: NextConfig = {
  rewrites: async () => [
    {
      source: "/api/chat/unified",
      destination: `${BACKEND_URL}/api/chat/unified`,
    },
    {
      source: "/api/apply",
      destination: `${BACKEND_URL}/api/apply`,
    },
    {
      source: "/api/resume/upload",
      destination: `${BACKEND_URL}/api/resume/upload`,
    },
  ],
};

export default nextConfig;
