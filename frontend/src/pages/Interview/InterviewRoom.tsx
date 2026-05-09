import { useEffect, useState, useCallback, useRef } from 'react';
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
import type { WSStageChangePayload, WSASRFinalPayload, WSLLMTokenPayload, WSErrorPayload } from '../../types/interview';

export const InterviewRoom = () => {
  const navigate = useNavigate();
  const { handleCreateInterview, handleFinishInterview } = useInterview();
  const { interviewId, stage, turns, setInterviewId, setStage, addTurn, updateLastTurn, setConnected } =
    useInterviewStore();
  const token = useAuthStore((s) => s.token) ?? '';

  const [wsError, setWsError] = useState<string | null>(null);

  // ── PCM 播放器（TTS 流式播放）─────────────────────────────────────────────
  const { enqueuePCM, flush: flushAudio } = usePCMPlayer();

  // ── WS 事件处理 ───────────────────────────────────────────────────────────
  const handleStageChange = useCallback(
    (payload: WSStageChangePayload) => {
      setStage(payload.stage === 'end' ? 'finished' : payload.stage);
      if (payload.stage === 'end') {
        navigate('/report/generating');
      }
    },
    [setStage, navigate],
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
    [stage, addTurn],
  );

  // AI 回复文字流：追加到最后一个 AI turn（question 字段）
  const llmTurnIdRef = useRef<string>('');
  const handleLLMToken = useCallback(
    (payload: WSLLMTokenPayload) => {
      if (llmTurnIdRef.current !== payload.turn_id) {
        // 新的 AI 回复轮次
        llmTurnIdRef.current = payload.turn_id;
        addTurn({
          turn_id: `ai_${payload.turn_id}`,
          stage,
          question: payload.token,
          user_answer: '',
        });
      } else {
        updateLastTurn({ question: turns.at(-1)?.question + payload.token });
      }
    },
    [stage, addTurn, updateLastTurn, turns],
  );

  const handleWSError = useCallback((payload: WSErrorPayload) => {
    setWsError(`错误 ${payload.code}：${payload.message}`);
  }, []);

  const handleReportReady = useCallback(() => {
    navigate('/report');
  }, [navigate]);

  // ── WebSocket ─────────────────────────────────────────────────────────────
  const effectiveInterviewId = interviewId ?? '';
  const {
    isConnected,
    connect,
    disconnect,
    sendAudioChunk,
    sendControl,
    sendCodeSubmit: _sendCodeSubmit,
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

  // ── 音频采集（AudioWorklet）────────────────────────────────────────────────
  const { isRecording, audioLevel, startRecording, stopRecording } = useAudio(
    effectiveInterviewId,
    { onPCMChunk: sendAudioChunk },
  );

  // ── 初始化：创建面试 → 建立 WS ────────────────────────────────────────────
  useEffect(() => {
    if (interviewId) {
      connect();
      return;
    }
    // 还没有 interviewId，先创建
    handleCreateInterview({ user_id: '' }).then((id) => {
      if (id) {
        setInterviewId(id);
        // interviewId 更新后 connect 会在下一轮 effect 触发
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [interviewId]);

  // interviewId 设定后连接 WS
  useEffect(() => {
    if (interviewId && !isConnected) {
      connect();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [interviewId]);

  // ── 录音开关 ──────────────────────────────────────────────────────────────
  const handleToggleRecording = useCallback(async () => {
    if (isRecording) {
      await stopRecording();
      sendControl('pause');
      flushAudio(); // 停止说话时打断 AI 播放（可根据产品策略调整）
    } else {
      const ok = await startRecording();
      if (ok) sendControl('start');
    }
  }, [isRecording, startRecording, stopRecording, sendControl, flushAudio]);

  // ── 结束面试 ──────────────────────────────────────────────────────────────
  const handleFinish = useCallback(async () => {
    if (!interviewId) return;
    await stopRecording();
    sendControl('stop');
    await handleFinishInterview(interviewId);
    disconnect();
    navigate('/questionnaire');
  }, [interviewId, stopRecording, sendControl, handleFinishInterview, disconnect, navigate]);

  return (
    <Container showHeader showFooter={false}>
      {/* 耳机提示横幅 */}
      <div className="bg-amber-50 border-b border-amber-200 px-4 py-2 flex items-center gap-2 text-sm text-amber-800">
        <svg className="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15.536 8.464a5 5 0 010 7.072M12 6a7 7 0 017 7v1a2 2 0 01-2 2h-1a2 2 0 01-2-2v-2a2 2 0 012-2h.93A7.001 7.001 0 005 13v1a2 2 0 01-2 2h-1a2 2 0 01-2-2v-2a2 2 0 012-2H3a7 7 0 017-7z"
          />
        </svg>
        建议佩戴耳机，以避免 AI 声音回声影响识别效果
      </div>

      <div className="h-[calc(100vh-7rem)] flex flex-col">
        {/* WS 错误提示 */}
        {wsError && (
          <div className="mx-4 mt-2 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {wsError}
            <button className="ml-2 underline" onClick={() => setWsError(null)}>
              关闭
            </button>
          </div>
        )}

        <div className="mb-4 px-4 pt-4">
          <StageProgress currentStage={stage} />
        </div>

        <div className="flex-1 grid grid-cols-1 lg:grid-cols-3 gap-4 min-h-0 px-4 pb-4">
          <div className="lg:col-span-2 flex flex-col">
            <ChatPanel turns={turns} />
          </div>

          <div className="flex flex-col gap-4">
            <div className="bg-white rounded-lg shadow-md p-4">
              <h3 className="font-semibold mb-3">音频控制</h3>
              <AudioIndicator isRecording={isRecording} audioLevel={audioLevel} />

              {/* 连接状态 */}
              <div className="flex items-center gap-1.5 mt-3">
                <span
                  className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-300'}`}
                />
                <span className="text-xs text-gray-500">
                  {isConnected ? '已连接' : '连接中…'}
                </span>
              </div>

              <Button
                className="mt-4 w-full"
                variant={isRecording ? 'secondary' : 'primary'}
                onClick={handleToggleRecording}
                disabled={!isConnected}
              >
                {isRecording ? '⏸ 暂停录音' : '🎤 开始说话'}
              </Button>

              <p className="text-xs text-gray-500 mt-2 text-center">
                {isRecording ? '正在录音，AI 实时转写中…' : '点击按钮开始回答'}
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
