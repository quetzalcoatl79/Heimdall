'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuthStore } from '@/lib/store/auth';
import { authApi } from '@/lib/api/auth';
import Link from 'next/link';
import Image from 'next/image';
import { useI18n } from '@/lib/i18n/I18nProvider';
import { useMemo } from 'react';

type LoginForm = {
  email: string;
  password: string;
};

export default function LoginPage() {
  const router = useRouter();
  const { setAuth } = useAuthStore();
  const { t } = useI18n();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const loginSchema = useMemo(
    () =>
      z.object({
        email: z.string().email(t('login.errors.email')),
        password: z.string().min(6, t('login.errors.password')),
      }),
    [t]
  );

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginForm) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await authApi.login(data.email, data.password);
      setAuth(response.user, response.access_token, response.refresh_token);
      router.push('/admin');
    } catch (err: any) {
      setError(err.response?.data?.error || t('login.errors.generic'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <main className="min-h-screen flex items-center justify-center p-4 bg-gray-50">
      <div className="w-full max-w-md">
        <div className="card">
          <div className="text-center mb-8">
            <div className="flex justify-center mb-4">
              <Image
                src="/Heimdall.png"
                alt="Logo Heimdall"
                width={96}
                height={120}
                className="drop-shadow"
                priority
              />
            </div>
            <h1 className="text-2xl font-bold">{t('login.title')}</h1>
            <p className="text-gray-600 mt-2">{t('login.subtitle')}</p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <div className="p-3 bg-red-100 text-red-700 rounded-lg text-sm">
                {error}
              </div>
            )}

            <div>
              <label htmlFor="email" className="block text-sm font-medium mb-1">
                {t('login.email')}
              </label>
              <input
                {...register('email')}
                type="email"
                id="email"
                className="input"
                placeholder="admin@heimdall.local"
              />
              {errors.email && (
                <p className="text-red-500 text-sm mt-1">{errors.email.message}</p>
              )}
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium mb-1">
                {t('login.password')}
              </label>
              <input
                {...register('password')}
                type="password"
                id="password"
                className="input"
                placeholder="••••••••"
              />
              {errors.password && (
                <p className="text-red-500 text-sm mt-1">{errors.password.message}</p>
              )}
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="btn btn-primary w-full"
            >
              {isLoading ? '...' : t('login.submit')}
            </button>
          </form>

          <div className="mt-6 text-center text-sm text-gray-600">
            <Link href="/" className="hover:underline">
              {t('login.back')}
            </Link>
          </div>
        </div>

        <div className="mt-4 text-center text-sm text-gray-500">
          <p>{t('login.defaultAccount')}</p>
        </div>
      </div>
    </main>
  );
}
