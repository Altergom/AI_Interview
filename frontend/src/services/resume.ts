import apiClient from './api';
import type { ApiResponse } from '../types/api';
import type {
  ResumeSubmitRequest,
  ResumeParseResponse,
  ResumeSubmitResponse,
} from '../types/resume';

// 解析简历 PDF
export const parseResume = async (file: File): Promise<ResumeParseResponse> => {
  const formData = new FormData();
  formData.append('file', file);

  const response = await apiClient.post<ApiResponse<ResumeParseResponse>>(
    '/resume/parse',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }
  );

  return response.data.data;
};

// 提交简历信息
export const submitResume = async (data: ResumeSubmitRequest): Promise<ResumeSubmitResponse> => {
  const response = await apiClient.post<ApiResponse<ResumeSubmitResponse>>('/resume/submit', data);
  return response.data.data;
};
