import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useInterviewStore } from '../store/interviewStore';

describe('useInterviewStore', () => {
  beforeEach(() => {
    const { reset } = useInterviewStore.getState();
    reset();
  });

  it('应该初始化为默认状态', () => {
    const { result } = renderHook(() => useInterviewStore());

    expect(result.current.interviewId).toBeNull();
    expect(result.current.stage).toBe('intro');
    expect(result.current.position).toBeNull();
    expect(result.current.direction).toBeNull();
    expect(result.current.turns).toEqual([]);
    expect(result.current.isConnected).toBe(false);
    expect(result.current.isProcessing).toBe(false);
  });

  it('应该设置面试 ID', () => {
    const { result } = renderHook(() => useInterviewStore());

    act(() => {
      result.current.setInterviewId('interview-123');
    });

    expect(result.current.interviewId).toBe('interview-123');
  });

  it('应该设置面试阶段', () => {
    const { result } = renderHook(() => useInterviewStore());

    act(() => {
      result.current.setStage('questioning');
    });

    expect(result.current.stage).toBe('questioning');
  });

  it('应该设置岗位和方向', () => {
    const { result } = renderHook(() => useInterviewStore());

    act(() => {
      result.current.setPosition('golang');
      result.current.setDirection('backend');
    });

    expect(result.current.position).toBe('golang');
    expect(result.current.direction).toBe('backend');
  });

  it('应该添加对话轮次', () => {
    const { result } = renderHook(() => useInterviewStore());
    const turn = {
      turn_id: 1,
      user_text: 'Hello',
      ai_text: 'Hi there',
      timestamp: new Date().toISOString(),
    };

    act(() => {
      result.current.addTurn(turn);
    });

    expect(result.current.turns).toHaveLength(1);
    expect(result.current.turns[0]).toEqual(turn);
  });

  it('应该更新最后一个对话轮次', () => {
    const { result } = renderHook(() => useInterviewStore());
    const turn = {
      turn_id: 1,
      user_text: 'Hello',
      ai_text: '',
      timestamp: new Date().toISOString(),
    };

    act(() => {
      result.current.addTurn(turn);
    });

    act(() => {
      result.current.updateLastTurn({ ai_text: 'Hi there' });
    });

    expect(result.current.turns[0].ai_text).toBe('Hi there');
  });

  it('应该设置连接状态', () => {
    const { result } = renderHook(() => useInterviewStore());

    act(() => {
      result.current.setConnected(true);
    });

    expect(result.current.isConnected).toBe(true);
  });

  it('应该重置状态', () => {
    const { result } = renderHook(() => useInterviewStore());

    act(() => {
      result.current.setInterviewId('test-id');
      result.current.setStage('questioning');
      result.current.setPosition('golang');
    });

    act(() => {
      result.current.reset();
    });

    expect(result.current.interviewId).toBeNull();
    expect(result.current.stage).toBe('intro');
    expect(result.current.position).toBeNull();
  });
});
