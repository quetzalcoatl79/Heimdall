'use client';

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { defaultLocale, messages, Locale } from './messages';

type I18nContextValue = {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: string, fallback?: string) => string;
};

const I18nContext = createContext<I18nContextValue | undefined>(undefined);

function getStoredLocale(): Locale {
  if (typeof window === 'undefined') return defaultLocale;
  const stored = window.localStorage.getItem('locale');
  return stored === 'en' ? 'en' : 'fr';
}

function getMessage(locale: Locale, key: string, fallback?: string): string {
  const parts = key.split('.');
  let current: any = messages[locale];
  for (const part of parts) {
    if (current && typeof current === 'object' && part in current) {
      current = current[part];
    } else {
      return fallback ?? key;
    }
  }
  if (typeof current === 'string') return current;
  return fallback ?? key;
}

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(defaultLocale);

  useEffect(() => {
    setLocaleState(getStoredLocale());
  }, []);

  const setLocale = useCallback((loc: Locale) => {
    setLocaleState(loc);
    if (typeof window !== 'undefined') {
      window.localStorage.setItem('locale', loc);
    }
  }, []);

  const t = useCallback((key: string, fallback?: string) => getMessage(locale, key, fallback), [locale]);

  const value = useMemo(() => ({ locale, setLocale, t }), [locale, setLocale, t]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18nContextValue {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useI18n must be used within I18nProvider');
  return ctx;
}
