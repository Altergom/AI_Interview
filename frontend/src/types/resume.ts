// 简历相关类型定义

// 教育背景
export interface ResumeEducation {
  school: string;
  major: string;
  graduation: string;
}

// 项目经验
export interface ResumeProject {
  name: string;
  tech_stack: string[];
  description: string;
  highlights: string[];
}

// 实习经历
export interface ResumeInternship {
  company?: string;
  role?: string;
  description?: string;
}

// 结构化简历
export interface StructuredResume {
  user_id: string;
  // 手动填表场景使用；PDF 解析路径目前不填充
  name?: string;
  phone?: string;
  email?: string;
  skills: string[];
  projects: ResumeProject[];
  internships: ResumeInternship[];
  education: ResumeEducation;
}

// 简历提交请求（表单方式）
export interface ResumeSubmitRequest {
  user_id: string;
  name?: string;
  phone?: string;
  email?: string;
  skills: string[];
  projects: ResumeProject[];
  internships: ResumeInternship[];
  education: ResumeEducation;
}

// 简历提交响应
export interface ResumeSubmitResponse {
  resume_id: string;
}

// 预签名上传地址响应
export interface ResumeUploadURLResponse {
  upload_url: string;
  object_key: string;
}

// 简历解析响应
export type ResumeParseResponse = StructuredResume;
