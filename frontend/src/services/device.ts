import apiClient from './api';
import type { ApiResponse } from '../types/api';

// 设备检测请求
export interface DeviceCheckRequest {
  has_microphone: boolean;
  has_camera: boolean;
  browser: string;
  os: string;
}

// 设备检测响应
export interface DeviceCheckResponse {
  status: string;
  message: string;
}

// 检测设备状态
export const checkDevice = async (data: DeviceCheckRequest): Promise<DeviceCheckResponse> => {
  const response = await apiClient.post<ApiResponse<DeviceCheckResponse>>('/device/check', data);
  return response.data.data;
};
