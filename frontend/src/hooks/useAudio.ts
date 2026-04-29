import { useState, useCallback, useRef, useEffect } from 'react';
import { submitAudio } from '../services/interview';

export const useAudio = (interviewId: string) => {
  const [isRecording, setIsRecording] = useState(false);
  const [audioLevel, setAudioLevel] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const turnIdRef = useRef<string>('');

  // 初始化音频上下文和分析器（用于音量检测）
  const initAudioAnalyser = useCallback((stream: MediaStream) => {
    const audioContext = new AudioContext();
    const analyser = audioContext.createAnalyser();
    const source = audioContext.createMediaStreamSource(stream);

    analyser.fftSize = 256;
    source.connect(analyser);

    audioContextRef.current = audioContext;
    analyserRef.current = analyser;

    // 实时检测音量
    const dataArray = new Uint8Array(analyser.frequencyBinCount);
    const detectVolume = () => {
      if (!analyserRef.current) return;

      analyserRef.current.getByteFrequencyData(dataArray);
      const average = dataArray.reduce((a, b) => a + b) / dataArray.length;
      setAudioLevel(average);

      if (isRecording) {
        requestAnimationFrame(detectVolume);
      }
    };

    detectVolume();
  }, [isRecording]);

  // 开始录音
  const startRecording = useCallback(async (): Promise<boolean> => {
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

      const mediaRecorder = new MediaRecorder(stream, {
        mimeType: 'audio/webm',
      });

      audioChunksRef.current = [];
      turnIdRef.current = `turn_${Date.now()}`;

      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunksRef.current.push(event.data);
        }
      };

      mediaRecorder.start();
      mediaRecorderRef.current = mediaRecorder;
      setIsRecording(true);

      initAudioAnalyser(stream);

      return true;
    } catch (err: any) {
      setError('录音启动失败');
      return false;
    }
  }, [initAudioAnalyser]);

  // 停止录音并上传
  const stopRecording = useCallback(async (): Promise<boolean> => {
    if (!mediaRecorderRef.current || !isRecording) return false;

    return new Promise((resolve) => {
      const mediaRecorder = mediaRecorderRef.current!;

      mediaRecorder.onstop = async () => {
        const audioBlob = new Blob(audioChunksRef.current, { type: 'audio/webm' });

        try {
          await submitAudio(audioBlob, {
            'X-Interview-Id': interviewId,
            'X-Turn-Id': turnIdRef.current,
          });

          resolve(true);
        } catch (err: any) {
          setError('音频上传失败');
          resolve(false);
        } finally {
          setIsRecording(false);
          audioChunksRef.current = [];
        }
      };

      mediaRecorder.stop();
      mediaRecorder.stream.getTracks().forEach(track => track.stop());
    });
  }, [interviewId, isRecording]);

  // 清理资源
  useEffect(() => {
    return () => {
      if (audioContextRef.current) {
        audioContextRef.current.close();
      }
      if (mediaRecorderRef.current && isRecording) {
        mediaRecorderRef.current.stop();
      }
    };
  }, [isRecording]);

  return {
    isRecording,
    audioLevel,
    error,
    startRecording,
    stopRecording,
  };
};

          });

          resolve(true);
        } catch (err: any) {
          setError('音频上传失败');
          resolve(false);
        } finally {
          setIsRecording(false);
