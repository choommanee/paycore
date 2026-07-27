/** @type {import('next').NextConfig} */
const backend = process.env.BACKEND_URL || "http://localhost:8080";
const nextConfig = {
  async rewrites() {
    return [
      { source: "/api/:path*", destination: `${backend}/v1/:path*` },
      // Public marketing landing: serve the static paycore.css site at the root.
      { source: "/", destination: "/index.html" },
    ];
  },
};
module.exports = nextConfig;
