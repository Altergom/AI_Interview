import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { useInterviewStore } from '../store/interviewStore';

// 模拟登录流程的集成测试
describe('Authentication Flow Integration', () => {
  beforeEach(() => {
    const { clearAuth } = useAuthStore.getState();
    clearAuth();
  });

  it('应该完成完整的登录流程', async () => {
    const user = userEvent.setup();

    // 模拟用户登录
    const mockUser = {
      user_id: '123',
      username: 'testuser',
      email: 'test@example.com',
      token: 'mock-token',
    };

    // 获取 store
    const { setUser, setToken } = useAuthStore.getState();

    // 执行登录
    setUser(mockUser);
    setToken('mock-token');

    // 验证状态
    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.user).toEqual(mockUser);
    expect(state.token).toBe('mock-token');
  });

  it('应该完成登出流程', async () => {
    // 先登录
    const mockUser = {
      user_id: '123',
      username: 'testuser',
      token: 'mock-token',
    };

    const { setUser, setToken, logout } = useAuthStore.getState();
    setUser(mockUser);
    setToken('mock-token');

    // 执行登出
    logout();

    // 验证状态已清空
    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.user).toBeNull();
    expect(state.token).toBeNull();
  });

  it('应该支持游客模式', () => {
    const { setGuest } = useAuthStore.getState();

    setGuest(true);

    const state = useAuthStore.getState();
    expect(state.isGuest).toBe(true);
    expect(state.isAuthenticated).toBe(true);
  });
});

// 模拟面试流程的集成测试
describe('Interview Flow Integration', () => {
  beforeEach(() => {
    const { reset } = useInterviewStore.getState();
    reset();
  });

  it('应该完成面试配置流程', () => {
    const { setPosition, setDirection, setInterviewId } = useInterviewStore.getState();

    // 设置岗位和方向
    setPosition('golang');
    setDirection('backend');
    setInterviewId('interview-123');

    const state = useInterviewStore.getState();
    expect(state.position).toBe('golang');
    expect(state.direction).toBe('backend');
    expect(state.interviewId).toBe('interview-123');
  });

  it('应该管理面试对话历史', () => {
    const { addTurn, updateLastTurn } = useInterviewStore.getState();

    // 添加对话
    const turn1 = {
      turn_id: 1,
      user_text: '你好',
      ai_text: '',
      timestamp: new Date().toISOString(),
    };

    addTurn(turn1);

    let state = useInterviewStore.getState();
    expect(state.turns).toHaveLength(1);

    // 更新最后一轮对话
    updateLastTurn({ ai_text: '你好，欢迎参加面试' });

    state = useInterviewStore.getState();
    expect(state.turns[0].ai_text).toBe('你好，欢迎参加面试');
  });
});
