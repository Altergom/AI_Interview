/**
 * useAudio — AudioWorklet 版本
 *
 * 替换旧的 MediaRecorder 方案。
 * - 采集：AudioWorklet（pcm-processor.js）输出 Int16 PCM 帧
 * - AEC：getUserMedia 约束开启 echoCancellation / noiseSuppression / autoGainControl
 * - 音量：AnalyserNode 实时检测（用于 UI 波形展示）
 * - 发送：每帧通过回调 onPCMChunk 传出，由 WS hook 发送二进制帧
 *
 * 不再依赖 submitAudio HTTP 接口。
 */
import { useState, useCallback, useRef, useEffect } from 'react';

interface UseAudioOptions {
  /** 接收到 PCM 帧时的回调（Int16 ArrayBuffer）*/
  onPCMChunk?: (chunk: ArrayBuffer) => void;
  /** 目标采样率，默认 16000 */
  targetSampleRate?: number;
}

export const useAudio = (_interviewId: string, options: UseAudioOptions = {}) => {
  const { onPCMChunk, targetSampleRate = 16000 } = options;

  const [isRecording, setIsRecording] = useState(false);
  const [audioLevel, setAudioLevel] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const audioContextRef = useRef<AudioContext | null>(null);
  const workletNodeRef = useRef<AudioWorkletNode | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const animFrameRef = useRef<number>(0);
  const onPCMChunkRef = useRef(onPCMChunk);

  // 保持回调引用最新，避免 stale closure
  useEffect(() => {
    onPCMChunkRef.current = onPCMChunk;
  }, [onPCMChunk]);

  // 音量检测循环
  const startVolumePoll = useCallback((analyser: AnalyserNode) => {
    const data = new Uint8Array(analyser.frequencyBinCount);
    const poll = () => {
      analyser.getByteFrequencyData(data);
      const avg = data.reduce((a, b) => a + b, 0) / data.length;
      setAudioLevel(avg);
      animFrameRef.current = requestAnimationFrame(poll);
    };
    animFrameRef.current = requestAnimationFrame(poll);
  }, []);

  const stopVolumePoll = useCallback(() => {
    if (animFrameRef.current) {
      cancelAnimationFrame(animFrameRef.current);
      animFrameRef.current = 0;
    }
    setAudioLevel(0);
  }, []);

  /** 开始录音 */
  const startRecording = useCallback(async (): Promise<boolean> => {
    setError(null);
    try {
      // AEC + 降噪 + 自动增益
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
          channelCount: 1,
          sampleRate: { ideal: targetSampleRate },
        },
      });
      streamRef.current = stream;

      const audioCtx = new AudioContext();
      audioContextRef.current = audioCtx;

      // 加载 AudioWorklet processor
      await audioCtx.audioWorklet.addModule('/pcm-processor.js');

      const workletNode = new AudioWorkletNode(audioCtx, 'pcm-processor', {
        processorOptions: { targetSampleRate },
      });
      workletNodeRef.current = workletNode;

      // 接收 PCM 帧 → 传给 WS
      workletNode.port.onmessage = (e: MessageEvent<ArrayBuffer>) => {
        if (onPCMChunkRef.current) {
          onPCMChunkRef.current(e.data);
        }
      };

      // 音量分析器
      const analyser = audioCtx.createAnalyser();
      analyser.fftSize = 256;
      analyserRef.current = analyser;

      const source = audioCtx.createMediaStreamSource(stream);
      source.connect(analyser);
      source.connect(workletNode);
      workletNode.connect(audioCtx.destination); // 必须连接到输出，否则 process() 不被调用

      startVolumePoll(analyser);
      setIsRecording(true);
      return true;
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '录音启动失败';
      setError(msg);
      return false;
    }
  }, [targetSampleRate, startVolumePoll]);

  /** 停止录音 */
  const stopRecording = useCallback(async (): Promise<void> => {
    stopVolumePoll();

    // 断开 Worklet
    if (workletNodeRef.current) {
      workletNodeRef.current.port.onmessage = null;
      workletNodeRef.current.disconnect();
      workletNodeRef.current = null;
    }

    // 关闭 AudioContext
    if (audioContextRef.current) {
      await audioContextRef.current.close().catch(() => {});
      audioContextRef.current = null;
    }

    // 停止媒体流轨道
    if (streamRef.current) {
      streamRef.current.getTracks().forEach(t => t.stop());
      streamRef.current = null;
    }

    setIsRecording(false);
  }, [stopVolumePoll]);

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      stopRecording();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return {
    isRecording,
    audioLevel,
    error,
    startRecording,
    stopRecording,
  };
};
