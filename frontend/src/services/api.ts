import axios from 'axios';
import type { AxiosInstance, AxiosError, InternalAxiosRequestConfig } from 'axios';
import type { ApiResponse } from '../types/api';
import { useAuthStore } from '../store/authStore';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/v1';

// 创建 Axios 实例
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器：添加 token
// 从 zustand store 读取，与登录/守卫保持唯一可信源，避免双存储不一致。
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().token;
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// 响应拦截器：统一错误处理
apiClient.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse | undefined;
    if (body && typeof body === 'object' && typeof body.code === 'number') {
      const msg = body.msg ?? body.message;
      if (body.code !== 0) {
        switch (body.code) {
          case 1401:
            useAuthStore.getState().clearAuth();
            window.location.href = '/login';
            break;
          default:
            console.error(`API Error [${body.code}]: ${msg}`);
        }

        return Promise.reject({
          ...body,
          msg,
        });
      }

      if (!body.msg && body.message) {
        body.msg = body.message;
      }
    }

    return response;
  },
  (error: AxiosError<ApiResponse>) => {
    if (error.response) {
      const { code } = error.response.data;
      const msg = error.response.data.msg ?? error.response.data.message;

      // 根据错误码处理
      // 1401 = CodeUnauthorized（与后端 internal/errors/code.go 对齐）
      switch (code) {
        case 1401:
          // token 无效或过期：清掉 zustand 鉴权状态后跳转登录
          useAuthStore.getState().clearAuth();
          window.location.href = '/login';
          break;
        default:
          console.error(`API Error [${code}]: ${msg}`);
      }

      return Promise.reject({
        ...error.response.data,
        msg,
      });
    }

    // 网络错误或超时
    return Promise.reject({
      code: -1,
      msg: error.message || '网络请求失败',
      data: null,
    });
  }
);

export default apiClient;
