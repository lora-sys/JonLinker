/** @type {import('next').NextConfig} */
const nextConfig = {
  typescript: {
    ignoreBuildErrors: false,
  },
  env: {
    NEXT_PUBLIC_WS_URL: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080',
  },
  images: {
    formats: ['image/avif', 'image/webp'],
    remotePatterns: [
      // Add specific image hosts here as needed
    ],
  },
  experimental: {
    optimizeCss: true,
  },
}

module.exports = nextConfig
