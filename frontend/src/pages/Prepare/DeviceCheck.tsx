import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../../components/layout/Container';
import { Button } from '../../components/common/Button';
import { Card } from '../../components/common/Card';
import { Loading } from '../../components/common/Loading';
import { MicrophoneTest } from './MicrophoneTest';
import { CameraTest } from './CameraTest';
import { useDevice } from '../../hooks/useDevice';
import { checkDevice } from '../../services/device';

export const DeviceCheck = () => {
  const navigate = useNavigate();
  const { requestPermissions, microphonePermission, cameraPermission } = useDevice();

  const [step, setStep] = useState<'permission' | 'microphone' | 'camera' | 'complete'>('permission');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleRequestPermissions = async () => {
    setLoading(true);
    setError('');

    try {
      await requestPermissions();
      setStep('microphone');
    } catch (err: any) {
      setError('无法获取设备权限，请检查浏览器设置');
    } finally {
      setLoading(false);
    }
  };

  const handleMicrophoneComplete = () => {
    setStep('camera');
  };

  const handleCameraComplete = () => {
    setStep('complete');
  };

  const handleStartInterview = async () => {
    setLoading(true);
    try {
      await checkDevice({
        has_microphone: microphonePermission === 'granted',
        has_camera: cameraPermission === 'granted',
        browser: navigator.userAgent,
        os: navigator.platform,
      });
      navigate('/interview');
    } catch (err) {
      setError('设备检测失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container showHeader>
      <div className="max-w-4xl mx-auto py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">设备检测</h1>
        <p className="text-gray-600 mb-4">
          请确保您的麦克风和摄像头工作正常
        </p>

        {/* 耳机提示横幅 */}
        <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-lg flex items-start gap-3">
          <svg className="w-5 h-5 text-amber-600 shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M15.536 8.464a5 5 0 010 7.072M12 6a7 7 0 017 7v1a2 2 0 01-2 2h-1a2 2 0 01-2-2v-2a2 2 0 012-2h.93A7.001 7.001 0 005 13v1a2 2 0 01-2 2h-1a2 2 0 01-2-2v-2a2 2 0 012-2H3a7 7 0 017-7z" />
          </svg>
          <div>
            <p className="text-amber-800 font-medium">建议佩戴耳机</p>
            <p className="text-amber-700 text-sm mt-0.5">
              无耳机时 AI 语音可能被麦克风拾取，影响识别准确度。已启用客户端回声消除（AEC），但耳机效果更佳。
            </p>
          </div>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            {error}
          </div>
        )}

        {step === 'permission' && (
          <Card>
            <div className="text-center py-8">
              <h3 className="text-xl font-semibold mb-4">请授予设备权限</h3>
              <p className="text-gray-600 mb-6">
                面试需要使用您的麦克风和摄像头，请点击下方按钮授权
              </p>
              <Button onClick={handleRequestPermissions} loading={loading}>
                授予权限
              </Button>
            </div>
          </Card>
        )}

        {step === 'microphone' && (
          <MicrophoneTest onComplete={handleMicrophoneComplete} />
        )}

        {step === 'camera' && (
          <CameraTest onComplete={handleCameraComplete} />
        )}

        {step === 'complete' && (
          <Card>
            <div className="text-center py-8">
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-4">设备检测完成</h3>
              <p className="text-gray-600 mb-6">
                您的设备已准备就绪，可以开始面试了
              </p>
              <Button onClick={handleStartInterview} loading={loading} size="lg">
                开始面试
              </Button>
            </div>
          </Card>
        )}
      </div>
    </Container>
  );
};
