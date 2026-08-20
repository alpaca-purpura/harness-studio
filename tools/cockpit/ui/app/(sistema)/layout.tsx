import { AppShell } from '@/components/layout/AppShell';

export default function SistemaLayout({ children }: { children: React.ReactNode }) {
  return <AppShell>{children}</AppShell>;
}
