import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAuthStore } from '../store/authStore';

describe('useAuthStore', () => {
  beforeEach(() => {
    const { clearAuth } = useAuthStore.getState();
    clearAuth();
  });

  it('应该初始化为未认证状态', () => {
    const { result } = renderHook(() => useAuthStore());

    expect(result.current.user).toBeNull();
    expect(result.current.token).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.isGuest).toBe(false);
  });

  it('应该设置用户信息', () => {
    const { result } = renderHook(() => useAuthStore());
    const mockUser = {
      user_id: '123',
      username: 'testuser',
      email: 'test@example.com',
      token: 'mock-token',
    };

    act(() => {
      result.current.setUser(mockUser);
    });

    expect(result.current.user).toEqual(mockUser);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it('应该设置 token', () => {
    const { result } = renderHook(() => useAuthStore());

    act(() => {
      result.current.setToken('test-token');
    });

    expect(result.current.token).toBe('test-token');
  });

  it('应该设置游客模式', () => {
    const { result } = renderHook(() => useAuthStore());

    act(() => {
      result.current.setGuest(true);
    });

    expect(result.current.isGuest).toBe(true);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it('应该登出并清空状态', () => {
    const { result } = renderHook(() => useAuthStore());
    const mockUser = {
      user_id: '123',
      username: 'testuser',
      token: 'mock-token',
    };

    act(() => {
      result.current.setUser(mockUser);
      result.current.setToken('test-token');
    });

    act(() => {
      result.current.logout();
    });

    expect(result.current.user).toBeNull();
    expect(result.current.token).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.isGuest).toBe(false);
  });
});
