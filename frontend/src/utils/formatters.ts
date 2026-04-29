import type { InterviewStage } from '../types/interview';

// 格式化时间戳为可读格式
export const formatDateTime = (timestamp: string): string => {
  const date = new Date(timestamp);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
};

// 格式化日期（仅日期部分）
export const formatDate = (timestamp: string): string => {
  const date = new Date(timestamp);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
};

// 格式化时长（秒 -> 分:秒）
export const formatDuration = (seconds: number): string => {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins}:${secs.toString().padStart(2, '0')}`;
};

// 格式化文件大小
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
};

// 格式化面试阶段为中文
export const formatStage = (stage: InterviewStage): string => {
  const stageMap: Record<InterviewStage, string> = {
    intro: '自我介绍',
    questioning: '技术提问',
    algorithm: '算法题',
    closing: '反问环节',
    finished: '面试结束',
  };
  return stageMap[stage] || stage;
};

// 格式化技能列表（数组 -> 逗号分隔字符串）
export const formatSkills = (skills: string[]): string => {
  return skills.join(', ');
};

// 格式化分数（保留1位小数）
export const formatScore = (score: number): string => {
  return score.toFixed(1);
};

// 截断文本（超长显示省略号）
export const truncateText = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + '...';
};
