'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

// Client-side redirect: funciona igual en dev (standalone) y en static export
// (el server redirect() de next/navigation no existe en output:'export').
export default function Home() {
  const router = useRouter();
  useEffect(() => {
    router.replace('/roadmap');
  }, [router]);
  return null;
}
