import { useState, useCallback, useRef } from 'react';
import { checkDevice } from '../services/device';

export const useDevice = () => {
  const [hasPermission, setHasPermission] = useState(false);
  const [hasMicrophone, setHasMicrophone] = useState(false);
  const [hasCamera, setHasCamera] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mediaStreamRef = useRef<MediaStream | null>(null);

  // 请求麦克风权限
  const requestMicrophonePermission = useCallback(async (): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      mediaStreamRef.current = stream;
      setHasMicrophone(true);
      setHasPermission(true);
      return true;
    } catch (err: any) {
      setError('麦克风权限获取失败');
      setHasMicrophone(false);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  // 请求摄像头权限
  const requestCameraPermission = useCallback(async (): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: true });
      if (mediaStreamRef.current) {
        stream.getTracks().forEach(track => mediaStreamRef.current?.addTrack(track));
      } else {
        mediaStreamRef.current = stream;
      }
      setHasCamera(true);
      setHasPermission(true);
      return true;
    } catch (err: any) {
      setError('摄像头权限获取失败');
      setHasCamera(false);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  // 提交设备检测结果到服务端
  const submitDeviceCheck = useCallback(async (): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      const browser = navigator.userAgent;
      const os = navigator.platform;

      await checkDevice({
        has_microphone: hasMicrophone,
        has_camera: hasCamera,
        browser,
        os,
      });

      return true;
    } catch (err: any) {
      setError('设备检测提交失败');
      return false;
    } finally {
      setLoading(false);
    }
  }, [hasMicrophone, hasCamera]);

  // 停止所有媒体流
  const stopMediaStream = useCallback(() => {
    if (mediaStreamRef.current) {
      mediaStreamRef.current.getTracks().forEach(track => track.stop());
      mediaStreamRef.current = null;
    }
    setHasMicrophone(false);
    setHasCamera(false);
    setHasPermission(false);
  }, []);

  return {
    hasPermission,
    hasMicrophone,
    hasCamera,
    loading,
    error,
    mediaStream: mediaStreamRef.current,
    requestMicrophonePermission,
    requestCameraPermission,
    submitDeviceCheck,
    stopMediaStream,
  };
};
