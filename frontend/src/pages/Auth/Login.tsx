import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Card } from '../../components/common/Card';
import { Input } from '../../components/common/Input';
import { Button } from '../../components/common/Button';
import { login } from '../../services/auth';
import { useAuthStore } from '../../store/authStore';
import { useToast } from '../../hooks/useToast';
import { validateEmail, validateRequired } from '../../utils/validators';

export const Login = () => {
  const navigate = useNavigate();
  const { setUser, setToken } = useAuthStore();
  const toast = useToast();

  const [formData, setFormData] = useState({
    email: '',
    password: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);

  const validate = () => {
    const newErrors: Record<string, string> = {};

    if (!validateRequired(formData.email)) {
      newErrors.email = '请输入邮箱';
    } else if (!validateEmail(formData.email)) {
      newErrors.email = '邮箱格式不正确';
    }

    if (!validateRequired(formData.password)) {
      newErrors.password = '请输入密码';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) return;

    setLoading(true);
    try {
      const response = await login(formData);
      setUser({
        user_id: response.user_id,
        username: response.username || '',
        email: formData.email,
        token: response.token,
      });
      setToken(response.token);
      toast.success('登录成功');
      navigate('/resume');
    } catch (error: any) {
      toast.error(error.response?.data?.message || '登录失败，请检查邮箱和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <Card className="w-full max-w-md fade-in">
        <h2 className="text-2xl font-bold text-center text-gray-900 mb-6">
          登录
        </h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="邮箱"
            type="email"
            value={formData.email}
            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
            error={errors.email}
            placeholder="请输入邮箱"
            autoComplete="email"
          />

          <Input
            label="密码"
            type="password"
            value={formData.password}
            onChange={(e) => setFormData({ ...formData, password: e.target.value })}
            error={errors.password}
            placeholder="请输入密码"
            autoComplete="current-password"
          />

          <Button
            type="submit"
            className="w-full"
            loading={loading}
            disabled={loading}
          >
            登录
          </Button>
        </form>

        <div className="mt-6 text-center text-sm text-gray-600">
          还没有账号？
          <Link to="/register" className="text-primary-600 hover:text-primary-700 ml-1 transition-colors">
            立即注册
          </Link>
        </div>

        <div className="mt-4 text-center">
          <Link
            to="/guest"
            className="text-sm text-gray-500 hover:text-gray-700 underline transition-colors"
          >
            游客模式体验
          </Link>
        </div>
      </Card>
    </div>
  );
};

