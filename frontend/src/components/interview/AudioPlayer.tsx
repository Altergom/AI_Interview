import { useRef, useEffect } from 'react';

interface AudioPlayerProps {
  audioBase64: string;
  autoPlay?: boolean;
  onEnded?: () => void;
}

export const AudioPlayer = ({ audioBase64, autoPlay = true, onEnded }: AudioPlayerProps) => {
  const audioRef = useRef<HTMLAudioElement>(null);

  useEffect(() => {
    if (!audioRef.current || !audioBase64) return;

    const audioBlob = base64ToBlob(audioBase64, 'audio/webm');
    const audioUrl = URL.createObjectURL(audioBlob);

    audioRef.current.src = audioUrl;

    if (autoPlay) {
      audioRef.current.play().catch(err => {
        console.error('Audio playback failed:', err);
      });
    }

    return () => {
      URL.revokeObjectURL(audioUrl);
    };
  }, [audioBase64, autoPlay]);

  const handleEnded = () => {
    if (onEnded) {
      onEnded();
    }
  };

  return (
    <audio
      ref={audioRef}
      onEnded={handleEnded}
      className="hidden"
    />
  );
};

// Base64 转 Blob
function base64ToBlob(base64: string, mimeType: string): Blob {
  const byteCharacters = atob(base64);
  const byteNumbers = new Array(byteCharacters.length);

  for (let i = 0; i < byteCharacters.length; i++) {
    byteNumbers[i] = byteCharacters.charCodeAt(i);
  }

  const byteArray = new Uint8Array(byteNumbers);
  return new Blob([byteArray], { type: mimeType });
}
