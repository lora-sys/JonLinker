import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '@/types';
import { apiClient } from '@/lib/api_client';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  _hasRehydrated: boolean;
  setAuth: (user: User, token: string) => void;
  clearAuth: () => void;
  updateUser: (user: Partial<User>) => void;
  setRehydrated: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      setAuth: (user, token) => {
        apiClient.setToken(token);
        apiClient.setUserId(user.id);
        set({ user, token, isAuthenticated: true });
      },

      clearAuth: () => {
        apiClient.setToken(null);
        set({ user: null, token: null, isAuthenticated: false });
      },

      updateUser: (userData) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...userData } : null,
        })),
    }),
    {
      name: 'joblinker-auth',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state?.token) {
          apiClient.setToken(state.token);
          if (state.user?.id) {
            apiClient.setUserId(state.user.id);
          }
        }
      },
    }
  )
);
