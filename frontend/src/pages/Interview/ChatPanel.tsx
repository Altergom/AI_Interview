import { useRef, useEffect } from 'react';
import type { InterviewTurn } from '../../types/interview';

interface ChatPanelProps {
  turns: InterviewTurn[];
}

export const ChatPanel = ({ turns }: ChatPanelProps) => {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [turns]);

  return (
    <div className="bg-white rounded-lg shadow-md p-4 flex-1 overflow-y-auto">
      <h3 className="font-semibold mb-4">对话记录</h3>

      <div className="space-y-4">
        {turns.map((turn, index) => (
          <div key={turn.turn_id || index}>
            <div className="flex justify-end mb-2">
              <div className="bg-primary-100 text-primary-900 rounded-lg px-4 py-2 max-w-[80%]">
                <p className="text-sm">{turn.user_answer}</p>
              </div>
            </div>

            <div className="flex justify-start">
              <div className="bg-gray-100 text-gray-900 rounded-lg px-4 py-2 max-w-[80%]">
                <p className="text-sm font-semibold text-primary-600 mb-1">AI 面试官</p>
                <p className="text-sm">{turn.question}</p>
              </div>
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
    </div>
  );
};
