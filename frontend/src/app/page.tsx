"use client";

import Image from 'next/image';
import Link from 'next/link';
import { useI18n } from '@/lib/i18n/I18nProvider';

export default function Home() {
  const { t } = useI18n();

  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-8">
      <div className="text-center flex flex-col items-center">
        <Image
          src="/Heimdall.png"
          alt="Logo Heimdall"
          width={160}
          height={200}
          className="mb-6 drop-shadow-lg"
          priority
        />
        <h1 className="text-4xl font-bold mb-4">{t('home.title')}</h1>
        <p className="text-xl text-gray-600 mb-8 max-w-2xl">{t('home.subtitle')}</p>
        
        <div className="flex gap-4 justify-center">
          <Link href="/login" className="btn btn-primary">
            {t('home.login')}
          </Link>
          <Link href="/admin" className="btn btn-secondary">
            {t('home.admin')}
          </Link>
        </div>
      </div>
      
      <div className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl">
        <div className="card">
          <h2 className="text-xl font-semibold mb-2">{t('home.cards.plugins.title')}</h2>
          <p className="text-gray-600">{t('home.cards.plugins.desc')}</p>
        </div>
        
        <div className="card">
          <h2 className="text-xl font-semibold mb-2">{t('home.cards.workers.title')}</h2>
          <p className="text-gray-600">{t('home.cards.workers.desc')}</p>
        </div>
        
        <div className="card">
          <h2 className="text-xl font-semibold mb-2">{t('home.cards.auth.title')}</h2>
          <p className="text-gray-600">{t('home.cards.auth.desc')}</p>
        </div>
      </div>
    </main>
  );
}
