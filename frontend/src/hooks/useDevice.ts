import { useCallback, useEffect, useRef } from 'react';
import { checkDevice } from '../services/device';
import { useDeviceStore } from '../store/deviceStore';

const getPermissionState = (granted: boolean): PermissionState => (
  granted ? 'granted' : 'denied'
);

export const useDevice = () => {
  const {
    microphonePermission,
    cameraPermission,
    microphoneDeviceId,
    cameraDeviceId,
    microphoneStream,
    cameraStream,
    isMicrophoneTested,
    isCameraTested,
    audioLevel,
    setMicrophonePermission,
    setCameraPermission,
    setMicrophoneDeviceId,
    setCameraDeviceId,
    setMicrophoneStream,
    setCameraStream,
    setMicrophoneTested,
    setCameraTested,
    setAudioLevel,
    stopAllStreams,
  } = useDeviceStore();

  const analyserRef = useRef<AnalyserNode | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const animationFrameRef = useRef<number>(0);

  const stopAudioLevelMonitor = useCallback(() => {
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
      animationFrameRef.current = 0;
    }
    analyserRef.current = null;

    if (audioContextRef.current) {
      audioContextRef.current.close().catch(() => {});
      audioContextRef.current = null;
    }

    setAudioLevel(0);
  }, [setAudioLevel]);

  const startAudioLevelMonitor = useCallback((stream: MediaStream) => {
    stopAudioLevelMonitor();

    const audioContext = new AudioContext();
    const analyser = audioContext.createAnalyser();
    analyser.fftSize = 256;

    const source = audioContext.createMediaStreamSource(stream);
    source.connect(analyser);

    audioContextRef.current = audioContext;
    analyserRef.current = analyser;

    const data = new Uint8Array(analyser.frequencyBinCount);
    const tick = () => {
      if (!analyserRef.current) {
        return;
      }

      analyserRef.current.getByteFrequencyData(data);
      const avg = data.reduce((sum, value) => sum + value, 0) / data.length;
      setAudioLevel(avg / 255);
      animationFrameRef.current = requestAnimationFrame(tick);
    };

    animationFrameRef.current = requestAnimationFrame(tick);
  }, [setAudioLevel, stopAudioLevelMonitor]);

  const applyMediaStream = useCallback((stream: MediaStream) => {
    const audioTracks = stream.getAudioTracks();
    const videoTracks = stream.getVideoTracks();

    const nextMicrophoneStream = audioTracks.length > 0 ? new MediaStream(audioTracks) : null;
    const nextCameraStream = videoTracks.length > 0 ? new MediaStream(videoTracks) : null;

    setMicrophoneStream(nextMicrophoneStream);
    setCameraStream(nextCameraStream);

    if (audioTracks[0]) {
      const settings = audioTracks[0].getSettings();
      setMicrophoneDeviceId(settings.deviceId ?? '');
      setMicrophonePermission('granted');
      startAudioLevelMonitor(nextMicrophoneStream!);
    } else {
      setMicrophoneDeviceId('');
      setMicrophonePermission('denied');
      stopAudioLevelMonitor();
    }

    if (videoTracks[0]) {
      const settings = videoTracks[0].getSettings();
      setCameraDeviceId(settings.deviceId ?? '');
      setCameraPermission('granted');
    } else {
      setCameraDeviceId('');
      setCameraPermission('denied');
    }
  }, [
    setCameraDeviceId,
    setCameraPermission,
    setCameraStream,
    setMicrophoneDeviceId,
    setMicrophonePermission,
    setMicrophoneStream,
    startAudioLevelMonitor,
    stopAudioLevelMonitor,
  ]);

  const requestPermissions = useCallback(async (): Promise<void> => {
    stopAllStreams();
    stopAudioLevelMonitor();

    const stream = await navigator.mediaDevices.getUserMedia({
      audio: {
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true,
      },
      video: true,
    });

    applyMediaStream(stream);
  }, [applyMediaStream, stopAllStreams, stopAudioLevelMonitor]);

  const requestMicrophonePermission = useCallback(async (): Promise<boolean> => {
    try {
      if (microphoneStream) {
        return true;
      }

      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const audioTracks = stream.getAudioTracks();
      const nextMicrophoneStream = audioTracks.length > 0 ? new MediaStream(audioTracks) : null;
      setMicrophoneStream(nextMicrophoneStream);
      setMicrophonePermission(getPermissionState(Boolean(nextMicrophoneStream)));

      if (audioTracks[0]) {
        const settings = audioTracks[0].getSettings();
        setMicrophoneDeviceId(settings.deviceId ?? '');
        startAudioLevelMonitor(nextMicrophoneStream!);
      }

      return Boolean(nextMicrophoneStream);
    } catch {
      setMicrophonePermission('denied');
      stopAudioLevelMonitor();
      return false;
    }
  }, [
    microphoneStream,
    setMicrophoneDeviceId,
    setMicrophonePermission,
    setMicrophoneStream,
    startAudioLevelMonitor,
    stopAudioLevelMonitor,
  ]);

  const requestCameraPermission = useCallback(async (): Promise<boolean> => {
    try {
      if (cameraStream) {
        return true;
      }

      const stream = await navigator.mediaDevices.getUserMedia({ video: true });
      const videoTracks = stream.getVideoTracks();
      const nextCameraStream = videoTracks.length > 0 ? new MediaStream(videoTracks) : null;
      setCameraStream(nextCameraStream);
      setCameraPermission(getPermissionState(Boolean(nextCameraStream)));

      if (videoTracks[0]) {
        const settings = videoTracks[0].getSettings();
        setCameraDeviceId(settings.deviceId ?? '');
      }

      return Boolean(nextCameraStream);
    } catch {
      setCameraPermission('denied');
      return false;
    }
  }, [cameraStream, setCameraDeviceId, setCameraPermission, setCameraStream]);

  const submitDeviceCheck = useCallback(async (): Promise<boolean> => {
    await checkDevice({
      has_microphone: microphonePermission === 'granted',
      has_camera: cameraPermission === 'granted',
      browser: navigator.userAgent,
      os: navigator.platform,
    });

    return true;
  }, [cameraPermission, microphonePermission]);

  useEffect(() => {
    return () => {
      stopAudioLevelMonitor();
    };
  }, [stopAudioLevelMonitor]);

  return {
    hasPermission: microphonePermission === 'granted' || cameraPermission === 'granted',
    hasMicrophone: microphonePermission === 'granted',
    hasCamera: cameraPermission === 'granted',
    loading: false,
    error: null,
    mediaStream: microphoneStream ?? cameraStream,
    microphonePermission,
    cameraPermission,
    microphoneDeviceId,
    cameraDeviceId,
    microphoneStream,
    cameraStream,
    isMicrophoneTested,
    isCameraTested,
    audioLevel,
    requestPermissions,
    requestMicrophonePermission,
    requestCameraPermission,
    submitDeviceCheck,
    setMicrophoneTested,
    setCameraTested,
    stopMediaStream: stopAllStreams,
  };
};
