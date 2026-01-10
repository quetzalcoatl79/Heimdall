'use client';

import { useI18n } from '@/lib/i18n/I18nProvider';
import { Globe } from 'lucide-react';

export function LanguageSwitcher({ variant = 'floating' }: { variant?: 'floating' | 'sidebar' }) {
  const { locale, setLocale, t } = useI18n();
  const next = locale === 'fr' ? 'en' : 'fr';

  if (variant === 'sidebar') {
    return (
      <button
        type="button"
        onClick={() => setLocale(next)}
        className="flex items-center gap-3 px-4 py-3 w-full text-gray-300 hover:bg-slate-800 rounded-lg transition-colors"
      >
        <Globe className="h-5 w-5" />
        {t('lang.toggle')}
      </button>
    );
  }

  return (
    <button
      type="button"
      onClick={() => setLocale(next)}
      className="fixed top-4 right-4 z-50 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-medium text-gray-700 shadow hover:bg-gray-50"
    >
      {t('lang.toggle')}
    </button>
  );
}
