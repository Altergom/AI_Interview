import { useNavigate } from 'react-router-dom';
import { Container } from '../components/layout/Container';
import { Button } from '../components/common/Button';
import { useInterviewStore } from '../store/interviewStore';
import { useAuthStore } from '../store/authStore';

export const EndPage = () => {
  const navigate = useNavigate();
  const { reset } = useInterviewStore();
  const { isGuest } = useAuthStore();

  const handleRestart = () => {
    reset();
    navigate('/resume');
  };

  return (
    <Container showHeader>
      <div className="max-w-2xl mx-auto text-center py-16">
        <div className="mb-8">
          <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-10 h-10 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="width" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-4">
            面试完成！
          </h1>
          <p className="text-lg text-gray-600">
            感谢您使用 AI 模拟面试系统，祝您求职顺利！
          </p>
        </div>

        {isGuest && (
          <div className="mb-8 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <p className="text-sm text-yellow-800">
              您当前使用的是游客模式，数据将在 24 小时后清理。
              <button
                onClick={() => navigate('/register')}
                className="text-yellow-900 underline ml-1"
              >
                立即注册
              </button>
              保存您的面试记录。
            </p>
          </div>
        )}

        <div className="space-y-4">
          <Button onClick={handleRestart} size="lg">
            再来一次
          </Button>
          <div>
            <button
              onClick={() => navigate('/')}
              className="text-gray-600 hover:text-gray-800"
            >
              返回首页
            </button>
          </div>
        </div>
      </div>
    </Container>
  );
};
