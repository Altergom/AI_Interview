import { useState, useCallback, useRef, useEffect } from 'react';
import type {
  SSEEventType,
  TextDeltaEvent,
  TextDoneEvent,
  AudioDeltaEvent,
  AudioDoneEvent,
  StageChangedEvent,
  CodeJudgedEvent,
  ReportReadyEvent,
  InterviewFinishedEvent,
  SSEErrorEvent,
} from '../types/interview';

const SSE_URL = import.meta.env.VITE_SSE_URL || 'http://localhost:8080/v1/interview/stream';

type SSEEventHandler = (data: any) => void;

export const useSSE = (interviewId: string) => {
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const eventSourceRef = useRef<EventSource | null>(null);
  const handlersRef = useRef<Map<SSEEventType, SSEEventHandler[]>>(new Map());
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);

  // 注册事件监听器
  const on = useCallback((eventType: SSEEventType, handler: SSEEventHandler) => {
    const handlers = handlersRef.current.get(eventType) || [];
    handlers.push(handler);
    handlersRef.current.set(eventType, handlers);
  }, []);

  // 移除事件监听器
  const off = useCallback((eventType: SSEEventType, handler: SSEEventHandler) => {
    const handlers = handlersRef.current.get(eventType) || [];
    const index = handlers.indexOf(handler);
    if (index > -1) {
      handlers.splice(index, 1);
      handlersRef.current.set(eventType, handlers);
    }
  }, []);

  // 触发事件
  const emit = useCallback((eventType: SSEEventType, data: any) => {
    const handlers = handlersRef.current.get(eventType) || [];
    handlers.forEach(handler => handler(data));
  }, []);

  // 连接 SSE
  const connect = useCallback(() => {
    if (eventSourceRef.current) return;

    const url = `${SSE_URL}?interview_id=${interviewId}`;
    const eventSource = new EventSource(url);

    eventSource.onopen = () => {
      setIsConnected(true);
      setError(null);
      reconnectAttemptsRef.current = 0;
    };

    eventSource.onerror = () => {
      setIsConnected(false);
      setError('SSE 连接失败');
      eventSource.close();
      eventSourceRef.current = null;

      // 断线重连（最多重试5次）
      if (reconnectAttemptsRef.current < 5) {
        reconnectAttemptsRef.current += 1;
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, 2000 * reconnectAttemptsRef.current);
      }
    };

    // 注册所有事件类型
    const eventTypes: SSEEventType[] = [
      'text.delta',
      'text.done',
      'audio.delta',
      'audio.done',
      'stage.changed',
      'code.judged',
      'resume.parsed',
      'report.ready',
      'interview.finished',
      'error',
    ];

    eventTypes.forEach(eventType => {
      eventSource.addEventListener(eventType, (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          emit(eventType, data);
        } catch (err) {
          console.error(`Failed to parse SSE event: ${eventType}`, err);
        }
      });
    });

    eventSourceRef.current = eventSource;
  }, [interviewId, emit]);

  // 断开连接
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    setIsConnected(false);
    reconnectAttemptsRef.current = 0;
  }, []);

  // 清理资源
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    error,
    connect,
    disconnect,
    on,
    off,
  };
};
