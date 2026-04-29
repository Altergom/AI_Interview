// 报告状态枚举
export type ReportStatus = 'pending' | 'generating' | 'done' | 'failed';

// 报告维度评分
export interface ReportDimensions {
  knowledge_depth: number;      // 知识深度 (0-10)
  expression: number;            // 表达能力 (0-10)
  problem_solving: number;       // 问题解决能力 (0-10)
  code_quality: number;          // 代码质量 (0-10)
  stress_response: number;       // 压力应对 (0-10)
}

// 报告状态查询响应
export interface ReportStatusResponse {
  status: ReportStatus;
  message: string;
}

// 报告详情响应
export interface ReportResponse {
  interview_id: string;
  dimensions: ReportDimensions;
  summary: string;                // 总体评价
  strong_points: string[];        // 优势点
  weak_points: string[];          // 待改进点
  created_at: string;
}

// 报告查询请求参数
export interface ReportQueryParams {
  interview_id: string;
}
