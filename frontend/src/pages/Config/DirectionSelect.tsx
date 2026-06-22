import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../../components/layout/Container';
import { Card } from '../../components/common/Card';
import { Button } from '../../components/common/Button';
import { DIRECTIONS } from '../../utils/constants';
import { useInterviewStore } from '../../store/interviewStore';
import { configInterview } from '../../services/interview';
import { useAuthStore } from '../../store/authStore';
import type { Direction } from '../../types/interview';

export const DirectionSelect = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { position, setDirection, setInterviewId } = useInterviewStore();
  const [selected, setSelected] = useState<Direction | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSelect = (direction: Direction) => {
    setSelected(direction);
    setDirection(direction);
  };

  const handleNext = async () => {
    if (!selected || !position || !user) return;

    setLoading(true);
    try {
      const response = await configInterview({
        user_id: user.user_id,
        position,
        direction: selected,
      });

      setInterviewId(response.interview_id);
      navigate('/prepare');
    } catch (error) {
      console.error('配置失败', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container showHeader>
      <div className="mx-auto max-w-4xl py-8">
        <h1 className="mb-2 text-3xl font-bold text-gray-900">选择面试方向</h1>
        <p className="mb-8 text-gray-600">请选择您的技术方向</p>

        <div className="mb-8 grid grid-cols-1 gap-4 md:grid-cols-2">
          {DIRECTIONS.map((dir) => (
            <Card
              key={dir.value}
              className={`cursor-pointer transition-all ${
                selected === dir.value
                  ? 'bg-primary-50 ring-2 ring-primary-500'
                  : 'hover:shadow-lg'
              }`}
              onClick={() => handleSelect(dir.value)}
            >
              <h3 className="mb-2 text-xl font-semibold text-gray-900">
                {dir.label}
              </h3>
              <p className="text-gray-600">{dir.description}</p>
            </Card>
          ))}
        </div>

        <div className="flex justify-between">
          <Button variant="secondary" onClick={() => navigate(-1)}>
            上一步
          </Button>
          <Button onClick={handleNext} disabled={!selected} loading={loading}>
            开始面试
          </Button>
        </div>
      </div>
    </Container>
  );
};
