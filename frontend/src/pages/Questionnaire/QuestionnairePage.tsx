import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../../components/layout/Container';
import { Button } from '../../components/common/Button';
import { Card } from '../../components/common/Card';
import { TurnItem } from './TurnItem';
import { getQuestionnaire, submitQuestionnaire } from '../../services/questionnaire';
import { useInterviewStore } from '../../store/interviewStore';
import type { InterviewTurn } from '../../types/interview';

export const QuestionnairePage = () => {
  const navigate = useNavigate();
  const { interviewId } = useInterviewStore();

  const [turns, setTurns] = useState<InterviewTurn[]>([]);
  const [feedback, setFeedback] = useState<Record<string, { quality: 'good' | 'bad'; feedback: string }>>({});
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (interviewId) {
      loadQuestionnaire();
    }
  }, [interviewId]);

  const loadQuestionnaire = async () => {
    if (!interviewId) return;

    try {
      const data = await getQuestionnaire(interviewId);
      setTurns(data.turns);
    } catch (error) {
      console.error('加载问卷失败', error);
    }
  };

  const handleFeedbackChange = (turnId: string, quality: 'good' | 'bad', feedbackText: string) => {
    setFeedback({
      ...feedback,
      [turnId]: { quality, feedback: feedbackText },
    });
  };

  const handleSubmit = async () => {
    if (!interviewId) return;

    setLoading(true);
    try {
      await submitQuestionnaire({
        interview_id: interviewId,
        answers: Object.entries(feedback).map(([turn_id, data]) => ({
          turn_id,
          quality: data.quality,
          feedback: data.feedback,
        })),
      });
      navigate('/report');
    } catch (error) {
      console.error('提交问卷失败', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container showHeader>
      <div className="max-w-4xl mx-auto py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">面试反馈</h1>
        <p className="text-gray-600 mb-8">
          请对本次面试的对话质量进行评价，帮助我们改进
        </p>

        <div className="space-y-6 mb-8">
          {turns.map((turn) => (
            <TurnItem
              key={turn.turn_id}
              turn={turn}
              onFeedbackChange={(quality, feedbackText) =>
                handleFeedbackChange(turn.turn_id, quality, feedbackText)
              }
            />
          ))}
        </div>

        <div className="flex justify-between">
          <Button variant="secondary" onClick={() => navigate('/report')}>
            跳过
          </Button>
          <Button onClick={handleSubmit} loading={loading}>
            提交反馈
          </Button>
        </div>
      </div>
    </Container>
  );
};
