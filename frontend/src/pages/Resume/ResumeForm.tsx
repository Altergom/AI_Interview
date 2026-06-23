import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../../components/layout/Container';
import { Button } from '../../components/common/Button';
import { BasicInfo } from './BasicInfo';
import { SkillsInput } from './SkillsInput';
import { ProjectForm } from './ProjectForm';
import { EducationForm } from './EducationForm';
import { PDFUpload } from './PDFUpload';
import { submitResume } from '../../services/resume';
import { useResumeStore } from '../../store/resumeStore';
import { useAuthStore } from '../../store/authStore';
import type { StructuredResume } from '../../types/resume';

export const ResumeForm = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { resume, setResume } = useResumeStore();

  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const steps = ['PDF上传', '基本信息', '技能', '项目经验', '教育背景'];

  const handleSubmit = async () => {
    if (!user || !resume) return;

    setLoading(true);
    setError('');

    try {
      await submitResume({
        user_id: user.user_id,
        ...resume,
      });
      navigate('/config');
    } catch (err: any) {
      setError(err?.msg || err?.response?.data?.msg || err?.response?.data?.message || err?.message || '提交失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    } else {
      handleSubmit();
    }
  };

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  return (
    <Container showHeader>
      <div className="max-w-4xl mx-auto py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">填写简历信息</h1>
        <p className="text-gray-600 mb-8">
          请填写您的简历信息，或上传 PDF 简历自动解析
        </p>

        {/* 步骤指示器 */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            {steps.map((step, index) => (
              <div key={index} className="flex items-center flex-1">
                <div className="flex flex-col items-center flex-1">
                  <div
                    className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold ${
                      index <= currentStep
                        ? 'bg-primary-500 text-white'
                        : 'bg-gray-200 text-gray-500'
                    }`}
                  >
                    {index + 1}
                  </div>
                  <span className="text-sm mt-2 text-gray-600">{step}</span>
                </div>
                {index < steps.length - 1 && (
                  <div
                    className={`h-1 flex-1 mx-2 ${
                      index < currentStep ? 'bg-primary-500' : 'bg-gray-200'
                    }`}
                  />
                )}
              </div>
            ))}
          </div>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            {error}
          </div>
        )}

        {/* 表单内容 */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          {currentStep === 0 && <PDFUpload />}
          {currentStep === 1 && <BasicInfo />}
          {currentStep === 2 && <SkillsInput />}
          {currentStep === 3 && <ProjectForm />}
          {currentStep === 4 && <EducationForm />}
        </div>

        {/* 操作按钮 */}
        <div className="flex justify-between">
          <Button
            variant="secondary"
            onClick={handlePrev}
            disabled={currentStep === 0}
          >
            上一步
          </Button>
          <Button
            onClick={handleNext}
            loading={loading}
            disabled={loading}
          >
            {currentStep === steps.length - 1 ? '提交' : '下一步'}
          </Button>
        </div>
      </div>
    </Container>
  );
};
