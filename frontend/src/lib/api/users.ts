import { apiClient } from './client';

interface UserListParams {
  page?: number;
  page_size?: number;
  role?: string;
  search?: string;
}

interface UserListResponse {
  users: User[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  is_active: boolean;
  last_login_at: string | null;
  created_at: string;
}

interface CreateUserInput {
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
  role?: string;
}

interface UpdateUserInput {
  first_name?: string;
  last_name?: string;
  role?: string;
  is_active?: boolean;
}

export const userApi = {
  list: async (params: UserListParams = {}): Promise<UserListResponse> => {
    const searchParams = new URLSearchParams();
    if (params.page) searchParams.set('page', String(params.page));
    if (params.page_size) searchParams.set('page_size', String(params.page_size));
    if (params.role) searchParams.set('role', params.role);
    if (params.search) searchParams.set('search', params.search);

    const response = await apiClient.get(`/admin/users?${searchParams.toString()}`);
    return response.data;
  },

  get: async (id: string): Promise<User> => {
    const response = await apiClient.get(`/admin/users/${id}`);
    return response.data;
  },

  create: async (data: CreateUserInput): Promise<User> => {
    const response = await apiClient.post('/admin/users', data);
    return response.data;
  },

  update: async (id: string, data: UpdateUserInput): Promise<User> => {
    const response = await apiClient.put(`/admin/users/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/admin/users/${id}`);
  },
};
