import { apiClient } from './client';

interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  user: {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    role: string;
  };
}

interface RegisterResponse {
  message: string;
  user: {
    id: string;
    email: string;
  };
}

export const authApi = {
  login: async (email: string, password: string): Promise<LoginResponse> => {
    const response = await apiClient.post('/auth/login', { email, password });
    return response.data;
  },

  register: async (
    email: string,
    password: string,
    firstName?: string,
    lastName?: string
  ): Promise<RegisterResponse> => {
    const response = await apiClient.post('/auth/register', {
      email,
      password,
      first_name: firstName,
      last_name: lastName,
    });
    return response.data;
  },

  refresh: async (refreshToken: string) => {
    const response = await apiClient.post('/auth/refresh', {
      refresh_token: refreshToken,
    });
    return response.data;
  },

  logout: async (refreshToken?: string) => {
    await apiClient.post('/auth/logout', { refresh_token: refreshToken });
  },

  me: async () => {
    const response = await apiClient.get('/users/me');
    return response.data;
  },
};
