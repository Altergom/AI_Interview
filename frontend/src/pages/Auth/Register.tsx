import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Card } from '../../components/common/Card';
import { Input } from '../../components/common/Input';
import { Button } from '../../components/common/Button';
import { register } from '../../services/auth';
import { useAuthStore } from '../../store/authStore';
import {
  validateEmail,
  validatePassword,
  validateUsername,
  validateRequired,
} from '../../utils/validators';

export const Register = () => {
  const navigate = useNavigate();
  const { setUser, setToken } = useAuthStore();

  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [apiError, setApiError] = useState('');

  const validate = () => {
    const newErrors: Record<string, string> = {};

    if (!validateRequired(formData.username)) {
      newErrors.username = '请输入用户名';
    } else if (!validateUsername(formData.username)) {
      newErrors.username = '用户名长度为2-20位，仅支持字母、数字、中文、下划线';
    }

    if (!validateRequired(formData.email)) {
      newErrors.email = '请输入邮箱';
    } else if (!validateEmail(formData.email)) {
      newErrors.email = '邮箱格式不正确';
    }

    if (!validateRequired(formData.password)) {
      newErrors.password = '请输入密码';
    } else if (!validatePassword(formData.password)) {
      newErrors.password = '密码至少8位，需包含字母和数字';
    }

    if (!validateRequired(formData.confirmPassword)) {
      newErrors.confirmPassword = '请确认密码';
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = '两次密码输入不一致';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setApiError('');

    if (!validate()) return;

    setLoading(true);
    try {
      const response = await register({
        username: formData.username,
        email: formData.email,
        password: formData.password,
      });
      setUser({
        user_id: response.user_id,
        username: formData.username,
        email: formData.email,
        token: response.token,
      });
      setToken(response.token);
      navigate('/resume');
    } catch (error: any) {
      setApiError(error?.msg || error?.response?.data?.msg || error?.response?.data?.message || error?.message || '注册失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <h2 className="text-2xl font-bold text-center text-gray-900 mb-6">
          注册
        </h2>

        {apiError && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {apiError}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="用户名"
            value={formData.username}
            onChange={(e) => setFormData({ ...formData, username: e.target.value })}
            error={errors.username}
            placeholder="请输入用户名"
          />

          <Input
            label="邮箱"
            type="email"
            value={formData.email}
            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
            error={errors.email}
            placeholder="请输入邮箱"
          />

          <Input
            label="密码"
            type="password"
            value={formData.password}
            onChange={(e) => setFormData({ ...formData, password: e.target.value })}
            error={errors.password}
            placeholder="至少8位，包含字母和数字"
          />

          <Input
            label="确认密码"
            type="password"
            value={formData.confirmPassword}
            onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
            error={errors.confirmPassword}
            placeholder="请再次输入密码"
          />

          <Button
            type="submit"
            className="w-full"
            loading={loading}
            disabled={loading}
          >
            注册
          </Button>
        </form>

        <div className="mt-6 text-center text-sm text-gray-600">
          已有账号？
          <Link to="/login" className="text-primary-600 hover:text-primary-700 ml-1">
            立即登录
          </Link>
        </div>
      </Card>
    </div>
  );
};
