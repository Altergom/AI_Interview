import type { Position, Direction, InterviewStage } from '../types/interview';

// 面试岗位选项
export const POSITIONS: Array<{ value: Position; label: string; description: string }> = [
  { value: 'golang', label: 'Golang开发', description: '后端开发，Go语言方向' },
  { value: 'java', label: 'Java开发', description: '后端开发，Java语言方向' },
  { value: 'frontend', label: '前端工程师', description: 'Web前端开发' },
  { value: 'test', label: '测试开发', description: '测试开发工程师' },
];

// 面试方向选项
export const DIRECTIONS: Array<{ value: Direction; label: string; description: string }> = [
  { value: 'backend', label: '软件开发', description: '传统后端开发方向' },
  { value: 'cloud', label: '云平台运维开发', description: '云原生、DevOps方向' },
  { value: 'agent', label: 'Agent开发', description: 'AI Agent开发方向' },
  { value: 'server', label: '服务端开发', description: '服务端架构方向' },
];

// 面试阶段配置
export const INTERVIEW_STAGES: Array<{
  value: InterviewStage;
  label: string;
  description: string;
}> = [
  { value: 'intro', label: '自我介绍', description: '候选人自我介绍阶段' },
  { value: 'questioning', label: '技术提问', description: '技术知识问答阶段' },
  { value: 'algorithm', label: '算法题', description: '算法编程阶段' },
  { value: 'closing', label: '反问环节', description: '候选人提问阶段' },
  { value: 'finished', label: '面试结束', description: '面试已结束' },
];

// API 基础地址
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/v1';

// SSE 连接地址
export const SSE_BASE_URL = import.meta.env.VITE_SSE_BASE_URL || 'http://localhost:8080/v1';

// localStorage 键名
export const STORAGE_KEYS = {
  TOKEN: 'ai_interview_token',
  USER_ID: 'ai_interview_user_id',
  USERNAME: 'ai_interview_username',
  INTERVIEW_ID: 'ai_interview_current_id',
  RESUME_DRAFT: 'ai_interview_resume_draft',
} as const;

// 文件上传限制
export const FILE_UPLOAD = {
  MAX_SIZE: 10 * 1024 * 1024, // 10MB
  ALLOWED_TYPES: ['application/pdf'],
  ALLOWED_EXTENSIONS: ['.pdf'],
} as const;

// 音频配置
export const AUDIO_CONFIG = {
  SAMPLE_RATE: 16000,
  CHANNELS: 1,
  BIT_DEPTH: 16,
  CHUNK_DURATION: 1000, // 1秒分片
} as const;
