import apiClient from './api';
import type { ApiResponse } from '../types/api';
import type { ReportStatusResponse, Report } from '../types/report';

// 查询报告状态
export const getReportStatus = async (interviewId: string): Promise<ReportStatusResponse> => {
  const response = await apiClient.get<ApiResponse<ReportStatusResponse>>(
    `/report/status?interview_id=${interviewId}`
  );
  return response.data.data;
};

// 获取报告
export const getReport = async (interviewId: string): Promise<Report> => {
  const response = await apiClient.get<ApiResponse<Report>>(
    `/report?interview_id=${interviewId}`
  );
  return response.data.data;
};
