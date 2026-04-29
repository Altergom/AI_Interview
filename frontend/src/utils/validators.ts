// 邮箱验证
export const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

// 密码验证（至少8位，包含字母和数字）
export const validatePassword = (password: string): boolean => {
  if (password.length < 8) return false;
  const hasLetter = /[a-zA-Z]/.test(password);
  const hasNumber = /\d/.test(password);
  return hasLetter && hasNumber;
};

// 用户名验证（2-20位，字母、数字、中文、下划线）
export const validateUsername = (username: string): boolean => {
  if (username.length < 2 || username.length > 20) return false;
  const usernameRegex = /^[一-龥a-zA-Z0-9_]+$/;
  return usernameRegex.test(username);
};

// 文件类型验证
export const validateFileType = (file: File, allowedTypes: string[]): boolean => {
  return allowedTypes.includes(file.type);
};

// 文件大小验证
export const validateFileSize = (file: File, maxSize: number): boolean => {
  return file.size <= maxSize;
};

// 必填字段验证
export const validateRequired = (value: string | null | undefined): boolean => {
  return value !== null && value !== undefined && value.trim() !== '';
};

// 技能列表验证（至少1项）
export const validateSkills = (skills: string[]): boolean => {
  return skills.length > 0 && skills.every(skill => skill.trim() !== '');
};

// 项目经验验证
export const validateProject = (project: {
  name: string;
  tech_stack: string[];
  description: string;
}): boolean => {
  return (
    validateRequired(project.name) &&
    project.tech_stack.length > 0 &&
    validateRequired(project.description)
  );
};

// 教育背景验证
export const validateEducation = (education: {
  school: string;
  major: string;
  graduation: string;
}): boolean => {
  return (
    validateRequired(education.school) &&
    validateRequired(education.major) &&
    validateRequired(education.graduation)
  );
};
