/**
 * useWebSocket — 面试 WebSocket 客户端
 *
 * 功能：
 * - 连接 WS，JWT token 通过 query param 传入（浏览器原生 WS 不支持自定义 header）
 * - 发送文本帧（控制指令、代码提交）
 * - 发送二进制帧（PCM 音频块）
 * - 接收文本帧 → 解析下行消息 → 回调分发
 * - 接收二进制帧 → onTTSAudio 回调（PCM 16kHz）
 * - 断线自动重连（指数退避，最多 5 次）
 */
import { useState, useCallback, useRef, useEffect } from 'react';
import type {
  WSDownMsg,
  WSUpMsg,
  WSControlPayload,
  WSCodeSubmitPayload,
  WSASRPartialPayload,
  WSASRFinalPayload,
  WSLLMTokenPayload,
  WSStageChangePayload,
  WSErrorPayload,
  WSReportReadyPayload,
} from '../types/interview';

const WS_BASE = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';

/** 所有下行事件回调，可选注册 */
export interface WSEventHandlers {
  onASRPartial?: (payload: WSASRPartialPayload) => void;
  onASRFinal?: (payload: WSASRFinalPayload) => void;
  onLLMToken?: (payload: WSLLMTokenPayload) => void;
  onTTSAudio?: (pcm: ArrayBuffer) => void;
  onStageChange?: (payload: WSStageChangePayload) => void;
  onError?: (payload: WSErrorPayload) => void;
  onReportReady?: (payload: WSReportReadyPayload) => void;
  onOpen?: () => void;
  onClose?: () => void;
}

const MAX_RECONNECT = 5;
const BASE_DELAY_MS = 1500;

export const useWebSocket = (
  interviewId: string,
  token: string,
  handlers: WSEventHandlers = {},
) => {
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const handlersRef = useRef(handlers);
  const reconnectCountRef = useRef(0);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // 主动关闭标记，避免触发重连
  const intentionalCloseRef = useRef(false);

  // 保持 handlers 引用最新
  useEffect(() => {
    handlersRef.current = handlers;
  });

  const clearReconnectTimer = () => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
  };

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    intentionalCloseRef.current = false;
    const url = `${WS_BASE}/v1/interview/ws/${interviewId}?token=${encodeURIComponent(token)}`;
    const ws = new WebSocket(url);
    ws.binaryType = 'arraybuffer';
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      setError(null);
      reconnectCountRef.current = 0;
      handlersRef.current.onOpen?.();
    };

    ws.onclose = () => {
      setIsConnected(false);
      handlersRef.current.onClose?.();

      if (intentionalCloseRef.current) return;

      // 指数退避重连
      if (reconnectCountRef.current < MAX_RECONNECT) {
        const delay = BASE_DELAY_MS * Math.pow(2, reconnectCountRef.current);
        reconnectCountRef.current += 1;
        reconnectTimerRef.current = setTimeout(() => connect(), delay);
      } else {
        setError('WebSocket 连接失败，请刷新页面重试');
      }
    };

    ws.onerror = () => {
      // onerror 后必然触发 onclose，在 onclose 中处理重连
      setError('WebSocket 连接出错');
    };

    ws.onmessage = (event: MessageEvent) => {
      // 二进制帧 → TTS PCM
      if (event.data instanceof ArrayBuffer) {
        handlersRef.current.onTTSAudio?.(event.data);
        return;
      }

      // 文本帧 → 下行消息
      try {
        const msg: WSDownMsg = JSON.parse(event.data as string);
        const h = handlersRef.current;

        switch (msg.type) {
          case 'asr_partial':
            h.onASRPartial?.(msg.payload as WSASRPartialPayload);
            break;
          case 'asr_final':
            h.onASRFinal?.(msg.payload as WSASRFinalPayload);
            break;
          case 'llm_token':
            h.onLLMToken?.(msg.payload as WSLLMTokenPayload);
            break;
          case 'stage_change':
            h.onStageChange?.(msg.payload as WSStageChangePayload);
            break;
          case 'error':
            h.onError?.(msg.payload as WSErrorPayload);
            break;
          case 'report_ready':
            h.onReportReady?.(msg.payload as WSReportReadyPayload);
            break;
          default:
            console.warn('[WS] unknown msg type:', msg.type);
        }
      } catch (e) {
        console.error('[WS] failed to parse message:', e);
      }
    };
  }, [interviewId, token]);

  const disconnect = useCallback(() => {
    intentionalCloseRef.current = true;
    clearReconnectTimer();
    if (wsRef.current) {
      wsRef.current.close(1000, 'user disconnect');
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  /** 发送 PCM 音频帧（二进制）*/
  const sendAudioChunk = useCallback((pcm: ArrayBuffer) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(pcm);
    }
  }, []);

  /** 发送控制指令 */
  const sendControl = useCallback((action: WSControlPayload['action']) => {
    const msg: WSUpMsg<WSControlPayload> = {
      type: 'control',
      payload: { action },
    };
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  /** 发送代码提交 */
  const sendCodeSubmit = useCallback((payload: WSCodeSubmitPayload) => {
    const msg: WSUpMsg<WSCodeSubmitPayload> = {
      type: 'code_submit',
      payload,
    };
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  // 组件卸载时断开
  useEffect(() => {
    return () => {
      intentionalCloseRef.current = true;
      clearReconnectTimer();
      wsRef.current?.close(1000, 'unmount');
    };
  }, []);

  return {
    isConnected,
    error,
    connect,
    disconnect,
    sendAudioChunk,
    sendControl,
    sendCodeSubmit,
  };
};
