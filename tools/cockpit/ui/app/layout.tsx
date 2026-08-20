import type { Metadata } from 'next';
import { getSelectableSistemas } from '@/lib/workspace';
import { SistemaProvider } from '@/components/providers/SistemaProvider';
import './globals.css';

export const metadata: Metadata = {
  title: 'Cockpit · SDD',
  description: 'Visualizador + editor del workflow Spec-Driven Development (filesystem-as-DB)',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  // En static export (binario Go) el layout se renderiza en build-time: no hay
  // workspace que leer. sistemas=[] → SistemaProvider las pide a GET /api/sistemas.
  const isStaticExport = process.env.COCKPIT_STATIC === '1';
  let sistemas: string[] = [];
  if (!isStaticExport) {
    try {
      sistemas = getSelectableSistemas();
    } catch {
      sistemas = ['main'];
    }
  }

  // Sistema por defecto del worktree (cockpit-up.sh exporta DEFAULT_SISTEMA={sistema}).
  // 'cross-sistema' (hub main) o vacío → sin preferencia, usa la primero detectado.
  const envSistema = process.env.DEFAULT_SISTEMA;
  const defaultSistema =
    envSistema && envSistema !== 'cross-sistema' ? envSistema : undefined;

  return (
    <html lang="es">
      <body>
        <SistemaProvider sistemas={sistemas} defaultSistema={defaultSistema}>
          {children}
        </SistemaProvider>
      </body>
    </html>
  );
}
