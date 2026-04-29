import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../../components/layout/Container';
import { Card } from '../../components/common/Card';
import { Button } from '../../components/common/Button';
import { POSITIONS } from '../../utils/constants';
import { useInterviewStore } from '../../store/interviewStore';
import type { Position } from '../../types/interview';

export const PositionSelect = () => {
  const navigate = useNavigate();
  const { setPosition } = useInterviewStore();
  const [selected, setSelected] = useState<Position | null>(null);

  const handleSelect = (position: Position) => {
    setSelected(position);
    setPosition(position);
  };

  const handleNext = () => {
    if (selected) {
      navigate('/config/direction');
    }
  };

  return (
    <Container showHeader>
      <div className="max-w-4xl mx-auto py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">选择面试岗位</h1>
        <p className="text-gray-600 mb-8">请选择您要面试的岗位</p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
          {POSITIONS.map((pos) => (
            <Card
              key={pos.value}
              className={`cursor-pointer transition-all ${
                selected === pos.value
                  ? 'ring-2 ring-primary-500 bg-primary-50'
                  : 'hover:shadow-lg'
              }`}
              onClick={() => handleSelect(pos.value)}
            >
              <h3 className="text-xl font-semibold text-gray-900 mb-2">
                {pos.label}
              </h3>
              <p className="text-gray-600">{pos.description}</p>
            </Card>
          ))}
        </div>

        <div className="flex justify-end">
          <Button onClick={handleNext} disabled={!selected}>
            下一步
          </Button>
        </div>
      </div>
    </Container>
  );
};
