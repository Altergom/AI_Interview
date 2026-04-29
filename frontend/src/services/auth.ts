import apiClient from './api';
import type { ApiResponse } from '../types/api';
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
} from '../types/user';

// 注册
export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/register', data);
  return response.data.data;
};

// 登录
export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/login', data);
  return response.data.data;
};

// 游客模式
export const guestLogin = async (): Promise<AuthResponse> => {
  const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/guest', {});
  return response.data.data;
};
