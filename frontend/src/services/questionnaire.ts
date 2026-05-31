import apiClient from './api';
import type { ApiResponse } from '../types/api';
import type { InterviewTurn } from '../types/interview';

// feedback 文本长度上限，与后端 maxFeedbackCharacters 保持一致。
// 后端超限会返回 BadRequest，前端用 maxLength 提前拦截。
export const MAX_FEEDBACK_LENGTH = 1000;

// 问卷答案
export interface QuestionnaireAnswer {
  turn_id: string;
  quality: 'good' | 'bad';
  feedback: string;
}

// 问卷提交请求
export interface QuestionnaireSubmitRequest {
  interview_id: string;
  answers: QuestionnaireAnswer[];
}

// 问卷提交响应（后端返回本次落库条目数）
export interface QuestionnaireSubmitResponse {
  submitted: number;
}

// 获取问卷：后端直接返回 InterviewTurn 扁平数组
export const getQuestionnaire = async (interviewId: string): Promise<InterviewTurn[]> => {
  const response = await apiClient.get<ApiResponse<InterviewTurn[]>>(
    `/questionnaire?interview_id=${interviewId}`
  );
  return response.data.data ?? [];
};

// 提交问卷
export const submitQuestionnaire = async (
  data: QuestionnaireSubmitRequest
): Promise<QuestionnaireSubmitResponse> => {
  const response = await apiClient.post<ApiResponse<QuestionnaireSubmitResponse>>(
    '/questionnaire/submit',
    data
  );
  return response.data.data;
};
