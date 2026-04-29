import { useNavigate } from 'react-router-dom';
import { Button } from '../components/common/Button';
import { useAuthStore } from '../store/authStore';

export const Index = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();

  const handleStart = () => {
    if (isAuthenticated) {
      navigate('/resume');
    } else {
      navigate('/login');
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-primary-100 flex items-center justify-center p-4 sm:p-6 lg:p-8">
      <div className="text-center max-w-2xl fade-in">
        <h1 className="text-3xl sm:text-4xl md:text-5xl font-bold text-primary-900 mb-4">
          AI 模拟面试系统
        </h1>
        <p className="text-base sm:text-lg md:text-xl text-gray-700 mb-8 px-4">
          智能面试官，真实面试体验，全方位评估你的技术能力
        </p>
        <div className="space-y-4">
          <Button
            size="lg"
            onClick={handleStart}
            className="px-8 sm:px-12 py-3 sm:py-4 text-base sm:text-lg w-full sm:w-auto"
          >
            开始面试
          </Button>
          {!isAuthenticated && (
            <div className="text-sm text-gray-600">
              <button
                onClick={() => navigate('/guest')}
                className="text-primary-600 hover:text-primary-700 underline transition-colors"
              >
                游客模式快速体验
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
