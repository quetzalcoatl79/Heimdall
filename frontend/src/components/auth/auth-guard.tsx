'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/lib/store/auth';
import Cookies from 'js-cookie';

interface AuthGuardProps {
  children: React.ReactNode;
  requiredRole?: string;
}

export function AuthGuard({ children, requiredRole }: AuthGuardProps) {
  const router = useRouter();
  const { isAuthenticated, user, isLoading } = useAuthStore();
  const [hydrated, setHydrated] = useState(false);

  // Attendre que le store soit hydraté
  useEffect(() => {
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    
    // Vérifier aussi le cookie directement
    const token = Cookies.get('access_token');
    const hasToken = !!token;
    
    if (!isLoading && !isAuthenticated && !hasToken) {
      router.push('/login');
      return;
    }

    if (requiredRole && user?.role !== requiredRole && user?.role !== 'admin') {
      router.push('/');
      return;
    }
  }, [isAuthenticated, isLoading, user, requiredRole, router, hydrated]);

  if (!hydrated || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  // Si on a un token, on laisse passer même si le store n'est pas encore synchro
  const token = Cookies.get('access_token');
  if (!isAuthenticated && !token) {
    return null;
  }

  if (requiredRole && user?.role !== requiredRole && user?.role !== 'admin') {
    return null;
  }

  return <>{children}</>;
}
