"use client";

import { useI18n } from '@/lib/i18n/I18nProvider';

export default function AdminSettingsPage() {
  const { t } = useI18n();
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t('admin.settings.title')}</h1>
      </div>

      <div className="card">
        <p className="text-gray-600">{t('admin.settings.wip')}</p>
      </div>
    </div>
  );
}
