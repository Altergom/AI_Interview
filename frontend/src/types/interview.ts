// 面试岗位枚举
export type Position = 'golang' | 'java' | 'frontend' | 'test';

// 面试方向枚举
export type Direction = 'backend' | 'cloud' | 'agent' | 'server';

// 面试阶段枚举
export type InterviewStage = 'intro' | 'questioning' | 'algorithm' | 'closing' | 'finished';

// 编程语言枚举
export type ProgrammingLanguage = 'java' | 'python' | 'go' | 'cpp';

// 面试配置请求
export interface InterviewConfigRequest {
  user_id: string;
  position: Position;
  direction: Direction;
}

// 面试配置响应
export interface InterviewConfigResponse {
  config_id: string;
  message: string;
}

// 创建面试请求
export interface CreateInterviewRequest {
  user_id: string;
}

// 创建面试响应
export interface CreateInterviewResponse {
  interview_id: string;
  stage: InterviewStage;
  created_at: string;
}

// 面试状态响应
export interface InterviewStateResponse {
  interview_id: string;
  stage: InterviewStage;
  questions_asked: number;
  current_question_followups: number;
  started_at: string;
}

// 音频提交请求头
export interface AudioSubmitHeaders {
  'X-Interview-Id': string;
  'X-Turn-Id': string;
}

// 音频提交响应
export interface AudioSubmitResponse {
  turn_id: string;
  status: string;
}

// 代码提交请求
export interface CodeSubmitRequest {
  interview_id: string;
  question_id: string;
  language: ProgrammingLanguage;
  code: string;
}

// 代码提交响应
export interface CodeSubmitResponse {
  status: string;
  message: string;
}

// 结束面试请求
export interface FinishInterviewRequest {
  interview_id: string;
}

// 结束面试响应
export interface FinishInterviewResponse {
  interview_id: string;
  finished_at: string;
  duration_seconds: number;
}

// 面试对话轮次
export interface InterviewTurn {
  turn_id: string;
  stage: InterviewStage;
  question: string;
  user_answer: string;
  asr_raw?: string;
  created_at?: string;
}

// SSE 事件类型
export type SSEEventType =
  | 'text.delta'
  | 'text.done'
  | 'audio.delta'
  | 'audio.done'
  | 'stage.changed'
  | 'code.judged'
  | 'resume.parsed'
  | 'report.ready'
  | 'interview.finished'
  | 'error';

// SSE 文字流增量事件
export interface TextDeltaEvent {
  turn_id: string;
  delta: string;
}

// SSE 文字流结束事件
export interface TextDoneEvent {
  turn_id: string;
  full_text: string;
}

// SSE 音频流增量事件
export interface AudioDeltaEvent {
  turn_id: string;
  audio_base64: string;
}

// SSE 音频流结束事件
export interface AudioDoneEvent {
  turn_id: string;
}

// SSE 阶段切换事件
export interface StageChangedEvent {
  from: InterviewStage;
  to: InterviewStage;
}

// SSE 代码评估完成事件
export interface CodeJudgedEvent {
  correctness: boolean;
  time_complexity: string;
  space_complexity: string;
  issues: string[];
}

// SSE 简历解析完成事件
export interface ResumeParsedEvent {
  status: string;
}

// SSE 报告生成完成事件
export interface ReportReadyEvent {
  interview_id: string;
}

// SSE 面试结束确认事件
export interface InterviewFinishedEvent {
  interview_id: string;
  finished_at: string;
}

// SSE 错误事件
export interface SSEErrorEvent {
  code: number;
  message: string;
}

// ─── WebSocket 消息类型（替换 SSE）────────────────────────────────────────────

// 上行消息类型
export type WSUpMsgType = 'control' | 'code_submit';

// 下行消息类型
export type WSDownMsgType =
  | 'asr_partial'
  | 'asr_final'
  | 'llm_token'
  | 'tts_audio'
  | 'stage_change'
  | 'error'
  | 'report_ready';

// 上行：通用包装（音频帧直接发二进制帧，不走此结构）
export interface WSUpMsg<T = unknown> {
  type: WSUpMsgType;
  payload?: T;
}

// 上行：控制指令
export interface WSControlPayload {
  /** start | pause | resume | stop */
  action: 'start' | 'pause' | 'resume' | 'stop';
}

// 上行：代码提交
export interface WSCodeSubmitPayload {
  language: string;
  code: string;
}

// 下行：通用包装
export interface WSDownMsg<T = unknown> {
  type: WSDownMsgType;
  payload?: T;
}

// 下行：ASR 中间结果
export interface WSASRPartialPayload {
  text: string;
}

// 下行：ASR 最终结果
export interface WSASRFinalPayload {
  text: string;
  turn_id: string;
}

// 下行：LLM 流式 token
export interface WSLLMTokenPayload {
  token: string;
  turn_id: string;
}

// 下行：阶段切换
export interface WSStageChangePayload {
  /** intro | questioning | closing | end */
  stage: InterviewStage | 'end';
  questions_asked: number;
}

// 下行：错误
export interface WSErrorPayload {
  code: number;
  message: string;
}

// 下行：报告就绪
export interface WSReportReadyPayload {
  interview_id: string;
}
