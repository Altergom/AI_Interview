// 用户类型
export interface User {
  user_id: string;
  username: string;
  email?: string;
  token: string;
  expires_at?: string;
}

// 登录请求
export interface LoginRequest {
  email: string;
  password: string;
}

// 注册请求
export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

// 认证响应
export interface AuthResponse {
  user_id: string;
  username?: string;
  token: string;
  expires_at?: string;
}
