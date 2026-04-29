import { Input } from '../../components/common/Input';
import { useResumeStore } from '../../store/resumeStore';

export const EducationForm = () => {
  const { resume, updateResume } = useResumeStore();
  const education = resume?.education || { school: '', major: '', degree: '', graduation: '' };

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">教育背景</h3>

      <Input
        label="学校"
        value={education.school}
        onChange={(e) => updateResume({ education: { ...education, school: e.target.value } })}
        placeholder="请输入学校名称"
      />

      <Input
        label="专业"
        value={education.major}
        onChange={(e) => updateResume({ education: { ...education, major: e.target.value } })}
        placeholder="请输入专业"
      />

      <Input
        label="学历"
        value={education.degree}
        onChange={(e) => updateResume({ education: { ...education, degree: e.target.value } })}
        placeholder="本科/硕士/博士"
      />

      <Input
        label="毕业时间"
        value={education.graduation}
        onChange={(e) => updateResume({ education: { ...education, graduation: e.target.value } })}
        placeholder="例如：2025-06"
      />
    </div>
  );
};
