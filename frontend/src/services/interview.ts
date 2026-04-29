import apiClient from './api';
import type { ApiResponse } from '../types/api';
import type {
  InterviewConfigRequest,
  InterviewConfigResponse,
  CreateInterviewRequest,
  CreateInterviewResponse,
  InterviewStateResponse,
  AudioSubmitHeaders,
  AudioSubmitResponse,
  CodeSubmitRequest,
  CodeSubmitResponse,
  FinishInterviewRequest,
  FinishInterviewResponse,
} from '../types/interview';

// 配置面试岗位和方向
export const configInterview = async (
  data: InterviewConfigRequest
): Promise<InterviewConfigResponse> => {
  const response = await apiClient.post<ApiResponse<InterviewConfigResponse>>(
    '/interview/config',
    data
  );
  return response.data.data;
};

// 创建面试
export const createInterview = async (
  data: CreateInterviewRequest
): Promise<CreateInterviewResponse> => {
  const response = await apiClient.post<ApiResponse<CreateInterviewResponse>>(
    '/interview/create',
    data
  );
  return response.data.data;
};

// 查询面试状态
export const getInterviewState = async (
  interviewId: string
): Promise<InterviewStateResponse> => {
  const response = await apiClient.get<ApiResponse<InterviewStateResponse>>(
    `/interview/state?interview_id=${interviewId}`
  );
  return response.data.data;
};

// 发送音频流
export const submitAudio = async (
  audioData: Blob,
  headers: AudioSubmitHeaders
): Promise<AudioSubmitResponse> => {
  const response = await apiClient.post<ApiResponse<AudioSubmitResponse>>(
    '/interview/audio',
    audioData,
    {
      headers: {
        'Content-Type': 'application/octet-stream',
        'X-Interview-Id': headers['X-Interview-Id'],
        'X-Turn-Id': headers['X-Turn-Id'],
      },
    }
  );
  return response.data.data;
};

// 提交代码
export const submitCode = async (data: CodeSubmitRequest): Promise<CodeSubmitResponse> => {
  const response = await apiClient.post<ApiResponse<CodeSubmitResponse>>(
    '/interview/code/submit',
    data
  );
  return response.data.data;
};

// 结束面试
export const finishInterview = async (
  data: FinishInterviewRequest
): Promise<FinishInterviewResponse> => {
  const response = await apiClient.post<ApiResponse<FinishInterviewResponse>>(
    '/interview/finish',
    data
  );
  return response.data.data;
};

