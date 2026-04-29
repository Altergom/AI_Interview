import { useState } from 'react';
import { Input } from '../../components/common/Input';
import { Button } from '../../components/common/Button';
import { useResumeStore } from '../../store/resumeStore';

export const SkillsInput = () => {
  const { resume, updateResume } = useResumeStore();
  const [inputValue, setInputValue] = useState('');

  const skills = resume?.skills || [];

  const handleAdd = () => {
    if (inputValue.trim()) {
      updateResume({ skills: [...skills, inputValue.trim()] });
      setInputValue('');
    }
  };

  const handleRemove = (index: number) => {
    updateResume({ skills: skills.filter((_, i) => i !== index) });
  };

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">技能列表</h3>

      <div className="flex gap-2">
        <Input
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          placeholder="输入技能后按回车或点击添加"
          onKeyPress={(e) => e.key === 'Enter' && handleAdd()}
        />
        <Button onClick={handleAdd}>添加</Button>
      </div>

      <div className="flex flex-wrap gap-2">
        {skills.map((skill, index) => (
          <span
            key={index}
            className="inline-flex items-center gap-1 px-3 py-1 bg-primary-100 text-primary-700 rounded-full text-sm"
          >
            {skill}
            <button
              onClick={() => handleRemove(index)}
              className="hover:text-primary-900"
            >
              ×
            </button>
          </span>
        ))}
      </div>
    </div>
  );
};
