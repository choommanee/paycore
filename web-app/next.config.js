/** @type {import('next').NextConfig} */
const backend = process.env.BACKEND_URL || "http://localhost:8080";
const nextConfig = {
  // Emit a self-contained server bundle for a lean production Docker image.
  output: "standalone",
  async rewrites() {
    return [
      { source: "/api/:path*", destination: `${backend}/v1/:path*` },
      // Public marketing landing: serve the static paycore.css site at the root.
      { source: "/", destination: "/index.html" },
    ];
  },
};
module.exports = nextConfig;
