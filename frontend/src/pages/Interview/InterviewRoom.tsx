import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../../components/layout/Container';
import { Button } from '../../components/common/Button';
import { ChatPanel } from './ChatPanel';
import { StageProgress } from './StageProgress';
import { AudioIndicator } from './AudioIndicator';
import { useInterview } from '../../hooks/useInterview';
import { useSSE } from '../../hooks/useSSE';
import { useAudio } from '../../hooks/useAudio';
import { useInterviewStore } from '../../store/interviewStore';

export const InterviewRoom = () => {
  const navigate = useNavigate();
  const { createInterview, finishInterview } = useInterview();
  const { interviewId, stage, turns } = useInterviewStore();
  const [isRecording, setIsRecording] = useState(false);

  useEffect(() => {
    if (!interviewId) {
      createInterview();
    }
  }, []);

  const handleFinish = async () => {
    if (interviewId) {
      await finishInterview(interviewId);
      navigate('/questionnaire');
    }
  };

  return (
    <Container showHeader showFooter={false}>
      <div className="h-[calc(100vh-4rem)] flex flex-col">
        <div className="mb-4">
          <StageProgress currentStage={stage} />
        </div>

        <div className="flex-1 grid grid-cols-1 lg:grid-cols-3 gap-4 min-h-0">
          <div className="lg:col-span-2 flex flex-col">
            <ChatPanel turns={turns} />
          </div>

          <div className="flex flex-col gap-4">
            <div className="bg-white rounded-lg shadow-md p-4">
              <h3 className="font-semibold mb-4">音频控制</h3>
              <AudioIndicator isRecording={isRecording} />
              <p className="text-sm text-gray-600 mt-4">
                {isRecording ? '正在录音...' : '等待您的回答'}
              </p>
            </div>

            <Button
              variant="secondary"
              onClick={handleFinish}
              className="w-full"
            >
              结束面试
            </Button>
          </div>
        </div>
      </div>
    </Container>
  );
};
