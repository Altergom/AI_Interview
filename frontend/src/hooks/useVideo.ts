import { useState, useCallback, useRef, useEffect } from 'react';

export const useVideo = () => {
  const [isStreaming, setIsStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const videoRef = useRef<HTMLVideoElement | null>(null);
  const streamRef = useRef<MediaStream | null>(null);

  // 开始视频流
  const startVideo = useCallback(async (videoElement: HTMLVideoElement): Promise<boolean> => {
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: {
          width: { ideal: 1280 },
          height: { ideal: 720 },
          facingMode: 'user',
        },
      });

      videoElement.srcObject = stream;
      videoRef.current = videoElement;
      streamRef.current = stream;
      setIsStreaming(true);

      return true;
    } catch (err: any) {
      setError('摄像头启动失败');
      return false;
    }
  }, []);

  // 停止视频流
  const stopVideo = useCallback(() => {
    if (streamRef.current) {
      streamRef.current.getTracks().forEach(track => track.stop());
      streamRef.current = null;
    }

    if (videoRef.current) {
      videoRef.current.srcObject = null;
      videoRef.current = null;
    }

    setIsStreaming(false);
  }, []);

  // 清理资源
  useEffect(() => {
    return () => {
      stopVideo();
    };
  }, [stopVideo]);

  return {
    isStreaming,
    error,
    startVideo,
    stopVideo,
  };
};
