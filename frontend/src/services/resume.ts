import apiClient from './api';
import type { ApiResponse } from '../types/api';
import type {
  ResumeSubmitRequest,
  ResumeParseResponse,
  ResumeSubmitResponse,
  ResumeUploadURLResponse,
} from '../types/resume';

const DEFAULT_PDF_CONTENT_TYPE = 'application/pdf';

// 解析简历 PDF：先拿预签名地址，再直传对象存储，最后用 object_key 触发解析
export const parseResume = async (file: File): Promise<ResumeParseResponse> => {
  const uploadURLResponse = await apiClient.get<ApiResponse<ResumeUploadURLResponse>>(
    '/resume/upload-url',
    {
      params: {
        filename: file.name,
      },
    }
  );

  const { upload_url, object_key } = uploadURLResponse.data.data;

  const uploadResponse = await fetch(upload_url, {
    method: 'PUT',
    headers: {
      'Content-Type': file.type || DEFAULT_PDF_CONTENT_TYPE,
    },
    body: file,
  });

  if (!uploadResponse.ok) {
    throw new Error(`简历上传失败 (${uploadResponse.status})`);
  }

  const response = await apiClient.post<ApiResponse<ResumeParseResponse>>(
    '/resume/parse',
    {
      object_key,
    }
  );

  return response.data.data;
};

// 提交简历信息
export const submitResume = async (data: ResumeSubmitRequest): Promise<ResumeSubmitResponse> => {
  const response = await apiClient.post<ApiResponse<ResumeSubmitResponse>>('/resume/submit', data);
  return response.data.data;
};
