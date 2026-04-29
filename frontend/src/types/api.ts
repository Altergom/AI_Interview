// API 响应基础类型
export interface ApiResponse<T = any> {
  code: number;
  data: T;
  message?: string;
}

// 分页响应
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}
