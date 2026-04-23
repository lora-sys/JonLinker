'use client';

import { useCallback } from 'react';
import { useAuthStore } from '@/stores/auth';
import type { LoginRequest, RegisterRequest } from '@/types/api';

export function useAuth() {
  const { user, token, isAuthenticated, setAuth, clearAuth } = useAuthStore();

  const login = useCallback(
    async (credentials: LoginRequest) => {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(credentials),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Login failed');
      }

      const data = await response.json();
      setAuth(data.user, data.token);
      return data;
    },
    [setAuth]
  );

  const register = useCallback(
    async (userData: RegisterRequest) => {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(userData),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Registration failed');
      }

      const data = await response.json();
      setAuth(data.user, data.token);
      return data;
    },
    [setAuth]
  );

  const logout = useCallback(() => {
    clearAuth();
  }, [clearAuth]);

  return {
    user,
    token,
    isAuthenticated,
    login,
    register,
    logout,
  };
}
