import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../../components/layout/Container';
import { Button } from '../../components/common/Button';
import { ChatPanel } from './ChatPanel';
import { StageProgress } from './StageProgress';
import { AudioIndicator } from './AudioIndicator';
import { useInterview } from '../../hooks/useInterview';
import { useWebSocket } from '../../hooks/useWebSocket';
import { useAudio } from '../../hooks/useAudio';
import { usePCMPlayer } from '../../hooks/usePCMPlayer';
import { useInterviewStore } from '../../store/interviewStore';
import { useAuthStore } from '../../store/authStore';
import type {
  WSErrorPayload,
  WSASRFinalPayload,
  WSLLMTokenPayload,
  WSStageChangePayload,
} from '../../types/interview';

export const InterviewRoom = () => {
  const navigate = useNavigate();
  const {
    handleCreateInterview,
    handleFinishInterview,
  } = useInterview();
  const {
    interviewId,
    stage,
    turns,
    setStage,
    addTurn,
    updateLastTurn,
    setConnected,
  } = useInterviewStore();
  const token = useAuthStore((s) => s.token) ?? '';

  const [wsError, setWsError] = useState<string | null>(null);
  const [createError, setCreateError] = useState<string | null>(null);
  const [isCreatingInterview, setIsCreatingInterview] = useState(false);

  const { enqueuePCM, flush: flushAudio } = usePCMPlayer();

  const handleStageChange = useCallback(
    (payload: WSStageChangePayload) => {
      setStage(payload.stage === 'end' ? 'finished' : payload.stage);
      if (payload.stage === 'end') {
        navigate('/report');
      }
    },
    [navigate, setStage],
  );

  const handleASRFinal = useCallback(
    (payload: WSASRFinalPayload) => {
      addTurn({
        turn_id: payload.turn_id,
        stage,
        question: '',
        user_answer: payload.text,
        asr_raw: payload.text,
      });
    },
    [addTurn, stage],
  );

  const llmTurnIdRef = useRef<string>('');
  const createdInterviewRef = useRef<string>('');
  const handleLLMToken = useCallback(
    (payload: WSLLMTokenPayload) => {
      if (llmTurnIdRef.current !== payload.turn_id) {
        llmTurnIdRef.current = payload.turn_id;
        addTurn({
          turn_id: `ai_${payload.turn_id}`,
          stage,
          question: payload.token,
          user_answer: '',
        });
        return;
      }

      const currentQuestion = turns.at(-1)?.question ?? '';
      updateLastTurn({ question: currentQuestion + payload.token });
    },
    [addTurn, stage, turns, updateLastTurn],
  );

  const handleWSError = useCallback((payload: WSErrorPayload) => {
    setWsError(`错误 ${payload.code}: ${payload.message}`);
  }, []);

  const handleReportReady = useCallback(() => {
    navigate('/report');
  }, [navigate]);

  const effectiveInterviewId = interviewId ?? '';
  const {
    isConnected,
    connect,
    disconnect,
    sendAudioChunk,
    sendControl,
  } = useWebSocket(effectiveInterviewId, token, {
    onTTSAudio: enqueuePCM,
    onStageChange: handleStageChange,
    onASRFinal: handleASRFinal,
    onLLMToken: handleLLMToken,
    onError: handleWSError,
    onReportReady: handleReportReady,
    onOpen: () => setConnected(true),
    onClose: () => setConnected(false),
  });

  const { isRecording, audioLevel, startRecording, stopRecording } = useAudio(
    effectiveInterviewId,
    { onPCMChunk: sendAudioChunk },
  );

  useEffect(() => {
    if (!interviewId) {
      setCreateError('缺少 interview_id，请返回上一页重新选择面试方向。');
      createdInterviewRef.current = '';
      return;
    }

    if (createdInterviewRef.current === interviewId || isCreatingInterview || isConnected) {
      return;
    }

    createdInterviewRef.current = interviewId;
    setIsCreatingInterview(true);
    setCreateError(null);

    handleCreateInterview({ interview_id: interviewId })
      .then((createdInterviewId) => {
        if (!createdInterviewId) {
          createdInterviewRef.current = '';
          setCreateError('创建面试失败，请返回上一页重新开始。');
          return;
        }

        connect();
      })
      .finally(() => {
        setIsCreatingInterview(false);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [interviewId]);

  const handleToggleRecording = useCallback(async () => {
    if (isRecording) {
      await stopRecording();
      sendControl('pause');
      flushAudio();
      return;
    }

    const ok = await startRecording();
    if (ok) {
      sendControl('start');
    }
  }, [flushAudio, isRecording, sendControl, startRecording, stopRecording]);

  const handleFinish = useCallback(async () => {
    if (!interviewId) return;

    await stopRecording();
    sendControl('stop');
    await handleFinishInterview(interviewId);
    disconnect();
    navigate('/questionnaire');
  }, [disconnect, handleFinishInterview, interviewId, navigate, sendControl, stopRecording]);

  return (
    <Container showHeader showFooter={false}>
      <div className="flex items-center gap-2 border-b border-amber-200 bg-amber-50 px-4 py-2 text-sm text-amber-800">
        <svg className="h-4 w-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15.536 8.464a5 5 0 010 7.072M12 6a7 7 0 017 7v1a2 2 0 01-2 2h-1a2 2 0 01-2-2v-2a2 2 0 012-2h.93A7.001 7.001 0 005 13v1a2 2 0 01-2 2h-1a2 2 0 01-2-2v-2a2 2 0 012-2H3a7 7 0 017-7z"
          />
        </svg>
        建议佩戴耳机，以避免 AI 语音回声影响识别效果
      </div>

      <div className="flex h-[calc(100vh-7rem)] flex-col">
        {createError && (
          <div className="mx-4 mt-2 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {createError}
          </div>
        )}

        {wsError && (
          <div className="mx-4 mt-2 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {wsError}
            <button className="ml-2 underline" onClick={() => setWsError(null)}>
              关闭
            </button>
          </div>
        )}

        <div className="mb-4 px-4 pt-4">
          <StageProgress currentStage={stage} />
        </div>

        <div className="grid min-h-0 flex-1 grid-cols-1 gap-4 px-4 pb-4 lg:grid-cols-3">
          <div className="flex flex-col lg:col-span-2">
            <ChatPanel turns={turns} />
          </div>

          <div className="flex flex-col gap-4">
            <div className="rounded-lg bg-white p-4 shadow-md">
              <h3 className="mb-3 font-semibold">音频控制</h3>
              <AudioIndicator isRecording={isRecording} audioLevel={audioLevel} />

              <div className="mt-3 flex items-center gap-1.5">
                <span
                  className={`h-2 w-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-300'}`}
                />
                <span className="text-xs text-gray-500">
                  {isConnected ? '已连接' : (isCreatingInterview ? '创建中...' : '连接中...')}
                </span>
              </div>

              <Button
                className="mt-4 w-full"
                variant={isRecording ? 'secondary' : 'primary'}
                onClick={handleToggleRecording}
                disabled={!isConnected}
              >
                {isRecording ? '暂停录音' : '开始说话'}
              </Button>

              <p className="mt-2 text-center text-xs text-gray-500">
                {isRecording ? '正在录音，AI 实时转写中...' : '点击按钮开始回答'}
              </p>
            </div>

            <Button variant="secondary" onClick={handleFinish} className="w-full">
              结束面试
            </Button>
          </div>
        </div>
      </div>
    </Container>
  );
};
