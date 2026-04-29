import { useState } from 'react';
import { Input } from '../../components/common/Input';
import { Button } from '../../components/common/Button';
import { useResumeStore } from '../../store/resumeStore';
import type { ResumeProject } from '../../types/resume';

export const ProjectForm = () => {
  const { resume, updateResume } = useResumeStore();
  const projects = resume?.projects || [];

  const [currentProject, setCurrentProject] = useState<Partial<ResumeProject>>({
    name: '',
    tech_stack: [],
    description: '',
    highlights: [],
  });

  const handleAdd = () => {
    if (currentProject.name && currentProject.description) {
      updateResume({
        projects: [...projects, currentProject as ResumeProject],
      });
      setCurrentProject({ name: '', tech_stack: [], description: '', highlights: [] });
    }
  };

  return (
    <div className="space-y-6">
      <h3 className="text-lg font-semibold text-gray-900">项目经验</h3>

      <div className="space-y-4 p-4 border border-gray-200 rounded-lg">
        <Input
          label="项目名称"
          value={currentProject.name || ''}
          onChange={(e) => setCurrentProject({ ...currentProject, name: e.target.value })}
          placeholder="请输入项目名称"
        />

        <Input
          label="项目描述"
          value={currentProject.description || ''}
          onChange={(e) => setCurrentProject({ ...currentProject, description: e.target.value })}
          placeholder="请描述项目内容"
        />

        <Button onClick={handleAdd} size="sm">
          添加项目
        </Button>
      </div>

      <div className="space-y-4">
        {projects.map((project, index) => (
          <div key={index} className="p-4 bg-gray-50 rounded-lg">
            <h4 className="font-semibold">{project.name}</h4>
            <p className="text-sm text-gray-600 mt-1">{project.description}</p>
          </div>
        ))}
      </div>
    </div>
  );
};
