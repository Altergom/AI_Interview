import { useState, useCallback } from 'react';
import {
  configInterview,
  createInterview,
  getInterviewState,
  submitCode,
  finishInterview,
} from '../services/interview';
import type {
  InterviewConfigRequest,
  CreateInterviewRequest,
  InterviewStateResponse,
  CodeSubmitRequest,
  InterviewStage,
} from '../types/interview';

export const useInterview = () => {
  const [interviewId, setInterviewId] = useState<string | null>(null);
  const [stage, setStage] = useState<InterviewStage>('intro');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 配置面试
  const handleConfigInterview = useCallback(async (
    data: InterviewConfigRequest
  ): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      await configInterview(data);
      return true;
    } catch (err: any) {
      setError(err?.msg || err?.message || '面试配置失败');
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  // 创建面试
  const handleCreateInterview = useCallback(async (
    data: CreateInterviewRequest
  ): Promise<string | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await createInterview(data);
      setInterviewId(response.interview_id);
      setStage(response.stage);
      return response.interview_id;
    } catch (err: any) {
      setError(err?.msg || err?.message || '创建面试失败');
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  // 查询面试状态
  const handleGetInterviewState = useCallback(async (
    id: string
  ): Promise<InterviewStateResponse | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await getInterviewState(id);
      setStage(response.stage);
      return response;
    } catch (err: any) {
      setError(err?.msg || err?.message || '查询面试状态失败');
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  // 提交代码
  const handleSubmitCode = useCallback(async (data: CodeSubmitRequest): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      await submitCode(data);
      return true;
    } catch (err: any) {
      setError(err?.msg || err?.message || '代码提交失败');
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  // 结束面试
  const handleFinishInterview = useCallback(async (id: string): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      await finishInterview({ interview_id: id });
      setStage('finished');
      return true;
    } catch (err: any) {
      setError(err?.msg || err?.message || '结束面试失败');
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    interviewId,
    stage,
    loading,
    error,
    setStage,
    handleConfigInterview,
    handleCreateInterview,
    handleGetInterviewState,
    handleSubmitCode,
    handleFinishInterview,
  };
};
