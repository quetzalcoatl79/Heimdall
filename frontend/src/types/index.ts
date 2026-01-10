// User types
export interface User {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role: 'admin' | 'worker' | 'user';
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

// Plugin types
export interface Plugin {
  id: string;
  name: string;
  version: string;
  description?: string;
  enabled: boolean;
  config: Record<string, unknown>;
  manifest: PluginManifest;
  installed_at: string;
}

export interface PluginManifest {
  routes?: PluginRoute[];
  permissions?: string[];
  menu_items?: PluginMenuItem[];
  hooks?: string[];
}

export interface PluginRoute {
  path: string;
  component: string;
}

export interface PluginMenuItem {
  label: string;
  path: string;
  icon?: string;
  position?: number;
}

// Job types
export interface Job {
  id: string;
  queue: string;
  type: string;
  payload: Record<string, unknown>;
  status: 'pending' | 'running' | 'completed' | 'failed';
  attempts: number;
  max_retries: number;
  error?: string;
  run_at?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

// API Response types
export interface ApiError {
  error: string;
  message?: string;
  code?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
