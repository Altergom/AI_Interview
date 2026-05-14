/**
 * usePCMPlayer — TTS PCM 流式播放
 *
 * 服务端推送的 tts_audio 是 16kHz / 16bit / mono PCM 二进制帧，
 * 本 hook 用 Web Audio API 拼接成连续播放队列，实现边收边播（低延迟）。
 *
 * 策略：
 * - 每帧 PCM 解码为 AudioBuffer，加入调度队列
 * - 调度器计算下一帧开始时间（context.currentTime + 已排队时长），
 *   保证帧与帧之间无缝衔接
 * - 调用 flush() 可清空队列（面试阶段切换时打断 AI 播放）
 */
import { useCallback, useRef, useEffect } from 'react';

const TTS_SAMPLE_RATE = 16000;

export const usePCMPlayer = () => {
  const audioCtxRef = useRef<AudioContext | null>(null);
  // 下一帧的计划开始时间（AudioContext 时间线）
  const nextStartTimeRef = useRef(0);

  /** 懒初始化 AudioContext（必须在用户手势后创建）*/
  const ensureAudioCtx = (): AudioContext => {
    if (!audioCtxRef.current || audioCtxRef.current.state === 'closed') {
      audioCtxRef.current = new AudioContext({ sampleRate: TTS_SAMPLE_RATE });
      nextStartTimeRef.current = 0;
    }
    if (audioCtxRef.current.state === 'suspended') {
      audioCtxRef.current.resume();
    }
    return audioCtxRef.current;
  };

  /**
   * 将服务端推来的 Int16 PCM ArrayBuffer 解码并排队播放。
   * 直接在 WebSocket onmessage 回调中调用即可。
   */
  const enqueuePCM = useCallback((buffer: ArrayBuffer) => {
    const ctx = ensureAudioCtx();

    // Int16 → Float32
    const int16 = new Int16Array(buffer);
    const float32 = new Float32Array(int16.length);
    for (let i = 0; i < int16.length; i++) {
      float32[i] = int16[i] / (int16[i] < 0 ? 0x8000 : 0x7fff);
    }

    // 创建 AudioBuffer
    const audioBuffer = ctx.createBuffer(1, float32.length, TTS_SAMPLE_RATE);
    audioBuffer.copyToChannel(float32, 0);

    // 计算此帧的开始时间
    const startAt = Math.max(ctx.currentTime, nextStartTimeRef.current);

    const source = ctx.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(ctx.destination);
    source.start(startAt);

    // 更新下一帧起始时间
    nextStartTimeRef.current = startAt + audioBuffer.duration;
  }, []);

  /**
   * 清空播放队列（打断当前 AI 音频，例如阶段切换或用户说话时）。
   * 通过重建 AudioContext 实现立即静音。
   */
  const flush = useCallback(() => {
    if (audioCtxRef.current) {
      audioCtxRef.current.close().catch(() => {});
      audioCtxRef.current = null;
    }
    nextStartTimeRef.current = 0;
  }, []);

  // 卸载时关闭 AudioContext
  useEffect(() => {
    return () => {
      audioCtxRef.current?.close().catch(() => {});
    };
  }, []);

  return { enqueuePCM, flush };
};
