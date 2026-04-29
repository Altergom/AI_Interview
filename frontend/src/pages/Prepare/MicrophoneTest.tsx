import { useEffect, useRef } from 'react';
import { Card } from '../../components/common/Card';
import { Button } from '../../components/common/Button';
import { AudioVisualizer } from './AudioVisualizer';
import { useDevice } from '../../hooks/useDevice';

interface MicrophoneTestProps {
  onComplete: () => void;
}

export const MicrophoneTest = ({ onComplete }: MicrophoneTestProps) => {
  const { microphoneStream, setMicrophoneTested, audioLevel } = useDevice();

  return (
    <Card>
      <h3 className="text-xl font-semibold mb-4">麦克风测试</h3>
      <p className="text-gray-600 mb-6">
        请对着麦克风说话，观察音频波形是否有变化
      </p>

      {microphoneStream && (
        <div className="mb-6">
          <AudioVisualizer stream={microphoneStream} />
          <div className="mt-4 text-center text-sm text-gray-600">
            当前音量：{Math.round(audioLevel * 100)}%
          </div>
        </div>
      )}

      <div className="flex justify-end">
        <Button onClick={() => { setMicrophoneTested(true); onComplete(); }}>
          麦克风正常，继续
        </Button>
      </div>
    </Card>
  );
};
