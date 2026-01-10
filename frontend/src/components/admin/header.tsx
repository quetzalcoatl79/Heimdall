'use client';

import { useAuthStore } from '@/lib/store/auth';
import { Bell, User } from 'lucide-react';

export function Header() {
  const { user } = useAuthStore();

  return (
    <header className="bg-white border-b border-gray-200 px-6 py-4">
      <div className="flex justify-end items-center">
        <div className="flex items-center gap-4">
          <button className="relative p-2 text-gray-400 hover:text-gray-600">
            <Bell className="h-5 w-5" />
            <span className="absolute top-1 right-1 h-2 w-2 bg-red-500 rounded-full"></span>
          </button>

          <div className="flex items-center gap-3 pl-4 border-l border-gray-200">
            <div className="h-8 w-8 bg-primary-100 rounded-full flex items-center justify-center">
              <User className="h-4 w-4 text-primary-600" />
            </div>
            <div className="text-sm">
              <p className="font-medium">
                {user?.first_name || user?.email || 'Admin'}
              </p>
              <p className="text-gray-500 text-xs">{user?.role}</p>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
