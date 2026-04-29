import { useEffect, useRef } from 'react';
import { Card } from '../../components/common/Card';
import { Button } from '../../components/common/Button';
import { useDevice } from '../../hooks/useDevice';

interface CameraTestProps {
  onComplete: () => void;
}

export const CameraTest = ({ onComplete }: CameraTestProps) => {
  const { cameraStream, setCameraTested } = useDevice();
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (cameraStream && videoRef.current) {
      videoRef.current.srcObject = cameraStream;
    }
  }, [cameraStream]);

  return (
    <Card>
      <h3 className="text-xl font-semibold mb-4">摄像头测试</h3>
      <p className="text-gray-600 mb-6">
        请确认能看到自己的画面
      </p>

      <div className="mb-6 bg-gray-900 rounded-lg overflow-hidden aspect-video">
        <video
          ref={videoRef}
          autoPlay
          playsInline
          muted
          className="w-full h-full object-cover"
        />
      </div>

      <div className="flex justify-end">
        <Button onClick={() => { setCameraTested(true); onComplete(); }}>
          摄像头正常，继续
        </Button>
      </div>
    </Card>
  );
};
