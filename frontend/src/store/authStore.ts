import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '../types/user';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isGuest: boolean;

  setUser: (user: User) => void;
  setToken: (token: string) => void;
  setGuest: (isGuest: boolean) => void;
  logout: () => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isGuest: false,

      setUser: (user) => set({
        user,
        isAuthenticated: true
      }),

      setToken: (token) => set({ token }),

      setGuest: (isGuest) => set({
        isGuest,
        isAuthenticated: isGuest
      }),

      logout: () => set({
        user: null,
        token: null,
        isAuthenticated: false,
        isGuest: false
      }),

      clearAuth: () => set({
        user: null,
        token: null,
        isAuthenticated: false,
        isGuest: false
      }),
    }),
    {
      name: 'auth-storage',
    }
  )
);
