import { useState, useCallback } from 'react';
import { login, register, guestLogin } from '../services/auth';
import type { LoginRequest, RegisterRequest, AuthResponse } from '../types/user';

export const useAuth = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 登录
  const handleLogin = useCallback(async (data: LoginRequest): Promise<AuthResponse | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await login(data);
      localStorage.setItem('token', response.token);
      localStorage.setItem('user', JSON.stringify(response));
      return response;
    } catch (err: any) {
      setError(err.message || '登录失败');
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  // 注册
  const handleRegister = useCallback(async (data: RegisterRequest): Promise<AuthResponse | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await register(data);
      localStorage.setItem('token', response.token);
      localStorage.setItem('user', JSON.stringify(response));
      return response;
    } catch (err: any) {
      setError(err.message || '注册失败');
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  // 游客模式
  const handleGuestLogin = useCallback(async (): Promise<AuthResponse | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await guestLogin();
      localStorage.setItem('token', response.token);
      localStorage.setItem('user', JSON.stringify(response));
      return response;
    } catch (err: any) {
      setError(err.message || '游客登录失败');
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  // 登出
  const handleLogout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  }, []);

  // 获取当前用户
  const getCurrentUser = useCallback((): AuthResponse | null => {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    try {
      return JSON.parse(userStr);
    } catch {
      return null;
    }
  }, []);

  // 检查是否已登录
  const isAuthenticated = useCallback((): boolean => {
    return !!localStorage.getItem('token');
  }, []);

  return {
    loading,
    error,
    handleLogin,
    handleRegister,
    handleGuestLogin,
    handleLogout,
    getCurrentUser,
    isAuthenticated,
  };
};
