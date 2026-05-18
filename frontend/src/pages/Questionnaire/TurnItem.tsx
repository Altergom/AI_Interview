import { useState } from 'react';
import { Card } from '../../components/common/Card';
import { formatStage } from '../../utils/formatters';
import { MAX_FEEDBACK_LENGTH } from '../../services/questionnaire';
import type { InterviewTurn } from '../../types/interview';

interface TurnItemProps {
  turn: InterviewTurn;
  onFeedbackChange: (quality: 'good' | 'bad', feedback: string) => void;
}

export const TurnItem = ({ turn, onFeedbackChange }: TurnItemProps) => {
  const [quality, setQuality] = useState<'good' | 'bad' | null>(null);
  const [feedback, setFeedback] = useState('');
  // 仅当 ASR 原文与最终采用的 user_answer 不同时才单独展示，避免视觉冗余
  const showAsrRaw = !!turn.asr_raw && turn.asr_raw !== turn.user_answer;

  const handleQualityChange = (newQuality: 'good' | 'bad') => {
    setQuality(newQuality);
    onFeedbackChange(newQuality, feedback);
  };

  const handleFeedbackChange = (newFeedback: string) => {
    setFeedback(newFeedback);
    if (quality) {
      onFeedbackChange(quality, newFeedback);
    }
  };

  return (
    <Card>
      <div className="mb-4">
        <span className="text-xs text-gray-500">{formatStage(turn.stage)}</span>
        <h4 className="font-semibold text-gray-900 mt-1">问题：{turn.question}</h4>
        <p className="text-sm text-gray-600 mt-2">您的回答：{turn.user_answer}</p>
        {showAsrRaw && (
          <p className="text-xs text-gray-400 mt-1">ASR 原文：{turn.asr_raw}</p>
        )}
      </div>

      <div className="space-y-3">
        <div>
          <p className="text-sm font-medium text-gray-700 mb-2">这轮对话质量如何？</p>
          <div className="flex gap-2">
            <button
              onClick={() => handleQualityChange('good')}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                quality === 'good'
                  ? 'bg-green-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              👍 很好
            </button>
            <button
              onClick={() => handleQualityChange('bad')}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                quality === 'bad'
                  ? 'bg-red-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              👎 不好
            </button>
          </div>
        </div>

        <div>
          <label className="text-sm font-medium text-gray-700 block mb-2">
            补充说明（可选）
          </label>
          <textarea
            value={feedback}
            onChange={(e) => handleFeedbackChange(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            rows={2}
            placeholder="请描述您的想法..."
            maxLength={MAX_FEEDBACK_LENGTH}
          />
          <div className="text-xs text-gray-400 text-right mt-1">
            {feedback.length}/{MAX_FEEDBACK_LENGTH}
          </div>
        </div>
      </div>
    </Card>
  );
};
