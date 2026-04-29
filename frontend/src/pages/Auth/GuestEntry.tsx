import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card } from '../../components/common/Card';
import { Button } from '../../components/common/Button';
import { guestLogin } from '../../services/auth';
import { useAuthStore } from '../../store/authStore';

export const GuestEntry = () => {
  const navigate = useNavigate();
  const { setUser, setToken, setGuest } = useAuthStore();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleGuestLogin = async () => {
    setError('');
    setLoading(true);

    try {
      const response = await guestLogin();
      setUser({
        user_id: response.user_id,
        username: '游客',
        token: response.token,
        expires_at: response.expires_at,
      });
      setToken(response.token);
      setGuest(true);
      navigate('/resume');
    } catch (error: any) {
      setError(error.response?.data?.message || '游客登录失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <h2 className="text-2xl font-bold text-center text-gray-900 mb-4">
          游客模式
        </h2>

        <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
          <h3 className="font-semibold text-yellow-900 mb-2">温馨提示</h3>
          <ul className="text-sm text-yellow-800 space-y-1">
            <li>• 游客账号有效期 24 小时</li>
            <li>• 面试数据将在 24 小时后自动清理</li>
            <li>• 可随时注册转为正式用户保存数据</li>
          </ul>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {error}
          </div>
        )}

        <Button
          onClick={handleGuestLogin}
          className="w-full mb-4"
          loading={loading}
          disabled={loading}
        >
          {loading ? '正在进入...' : '以游客身份开始'}
        </Button>

        <div className="text-center">
          <button
            onClick={() => navigate('/login')}
            className="text-sm text-gray-600 hover:text-gray-800"
          >
            返回登录
          </button>
        </div>
      </Card>
    </div>
  );
};
