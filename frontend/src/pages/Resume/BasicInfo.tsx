import { Input } from '../../components/common/Input';
import { useResumeStore } from '../../store/resumeStore';

export const BasicInfo = () => {
  const { resume, updateResume } = useResumeStore();

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">基本信息</h3>

      <Input
        label="姓名"
        value={resume?.name || ''}
        onChange={(e) => updateResume({ name: e.target.value })}
        placeholder="请输入姓名"
      />

      <Input
        label="联系电话"
        value={resume?.phone || ''}
        onChange={(e) => updateResume({ phone: e.target.value })}
        placeholder="请输入联系电话"
      />

      <Input
        label="邮箱"
        type="email"
        value={resume?.email || ''}
        onChange={(e) => updateResume({ email: e.target.value })}
        placeholder="请输入邮箱"
      />
    </div>
  );
};
