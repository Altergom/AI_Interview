import { INTERVIEW_STAGES } from '../../utils/constants';
import type { InterviewStage } from '../../types/interview';

interface StageProgressProps {
  currentStage: InterviewStage;
}

export const StageProgress = ({ currentStage }: StageProgressProps) => {
  const currentIndex = INTERVIEW_STAGES.findIndex(s => s.value === currentStage);

  return (
    <div className="bg-white rounded-lg shadow-md p-4">
      <h3 className="font-semibold mb-4">面试进度</h3>
      <div className="flex items-center justify-between">
        {INTERVIEW_STAGES.filter(s => s.value !== 'finished').map((stage, index) => (
          <div key={stage.value} className="flex items-center flex-1">
            <div className="flex flex-col items-center flex-1">
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold ${
                  index <= currentIndex
                    ? 'bg-primary-500 text-white'
                    : 'bg-gray-200 text-gray-500'
                }`}
              >
                {index + 1}
              </div>
              <span className="text-xs mt-1 text-gray-600">{stage.label}</span>
            </div>
            {index < INTERVIEW_STAGES.length - 2 && (
              <div
                className={`h-1 flex-1 mx-2 ${
                  index < currentIndex ? 'bg-primary-500' : 'bg-gray-200'
                }`}
              />
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
