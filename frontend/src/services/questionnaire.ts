import apiClient from './api';
import type { ApiResponse } from '../types/api';

// 问卷数据
export interface QuestionnaireData {
  interview_id: string;
  turns: QuestionnaireTurn[];
}

// 问卷单轮对话
export interface QuestionnaireTurn {
  turn_id: string;
  stage: string;
  question: string;
  user_answer: string;
}

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

// 问卷提交响应
export interface QuestionnaireSubmitResponse {
  message: string;
}

// 获取问卷
export const getQuestionnaire = async (interviewId: string): Promise<QuestionnaireData> => {
  const response = await apiClient.get<ApiResponse<QuestionnaireData>>(
    `/questionnaire?interview_id=${interviewId}`
  );
  return response.data.data;
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
