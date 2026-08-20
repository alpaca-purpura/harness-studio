/** @type {import('next').NextConfig} */
// COCKPIT_STATIC=1 → output:'export' (UI estático embebido en el binario Go).
// Sin la env → 'standalone' (dev / despliegue Node como hasta ahora).
const isStatic = process.env.COCKPIT_STATIC === '1';

const nextConfig = {
  output: isStatic ? 'export' : 'standalone',
  ...(isStatic ? { trailingSlash: true } : {}),
  reactStrictMode: true,
  // El cockpit lee filesystem fuera de su raíz (WORKSPACE_ROOT) · marcamos paquetes server externos para no bundlear
  serverExternalPackages: ['simple-git', 'chokidar'],
};

module.exports = nextConfig;
