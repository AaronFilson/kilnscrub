/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "export",
  images: {
    unoptimized: true,
  },
  trailingSlash: false,
  poweredByHeader: false,
  productionBrowserSourceMaps: false,
};

export default nextConfig;
