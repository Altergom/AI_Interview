package respx

type ErrorCode int

const (
	CodeOK           ErrorCode = 0
	CodeBadRequest   ErrorCode = 1400
	CodeUnauthorized ErrorCode = 1401
	CodeForbidden    ErrorCode = 1403
	CodeNotFound     ErrorCode = 1404
	CodeInternal     ErrorCode = 1500

	CodeEmailRegistered ErrorCode = 1101
	CodeUserNotFound    ErrorCode = 1102
	CodeWrongPassword   ErrorCode = 1103

	CodeResumeNotFound    ErrorCode = 2001
	CodeResumeParseFailed ErrorCode = 2002
	CodeResumeDuplicate   ErrorCode = 2003

	CodeInterviewSessionNotFound ErrorCode = 3001
	CodeInterviewStageInvalid    ErrorCode = 3002
	CodeInterviewForbidden       ErrorCode = 3003
	CodeInterviewTurnNotFound    ErrorCode = 3004

	CodeStorageUploadFailed ErrorCode = 4001
	CodeStorageNotFound     ErrorCode = 4002

	CodeExportPDFFailed ErrorCode = 5001

	CodeKnowledgeBaseNotFound ErrorCode = 6001
	CodeVectorIndexFailed     ErrorCode = 6002

	CodeAIServiceTimeout         ErrorCode = 7002
	CodeAIStructuredOutputFailed ErrorCode = 7003
	CodeAIProviderTestFailed     ErrorCode = 7004
	CodeAIProviderNotFound       ErrorCode = 7005

	CodeRateLimitExceeded ErrorCode = 8001

	CodeInterviewScheduleNotFound ErrorCode = 9001
	CodeScheduleParseFailed       ErrorCode = 9002

	CodeVoiceSessionNotFound ErrorCode = 10001
	CodeWSConnectionFailed   ErrorCode = 10002
)

var defaultMessages = map[ErrorCode]string{
	CodeOK:           "成功",
	CodeBadRequest:   "请求参数错误",
	CodeUnauthorized: "未授权，请先登录",
	CodeForbidden:    "无权限",
	CodeNotFound:     "资源不存在",
	CodeInternal:     "服务内部错误",

	CodeEmailRegistered: "邮箱已注册",
	CodeUserNotFound:    "用户不存在",
	CodeWrongPassword:   "密码错误",

	CodeResumeNotFound:    "简历不存在",
	CodeResumeParseFailed: "简历解析失败",
	CodeResumeDuplicate:   "简历已存在，直接返回解析结果",

	CodeInterviewSessionNotFound: "面试会话不存在",
	CodeInterviewStageInvalid:    "当前阶段不支持此操作",
	CodeInterviewForbidden:       "无权访问该面试",
	CodeInterviewTurnNotFound:    "面试 turn 不存在",

	CodeStorageUploadFailed: "文件上传失败",
	CodeStorageNotFound:     "文件不存在",

	CodeExportPDFFailed: "PDF 导出失败",

	CodeKnowledgeBaseNotFound: "知识库不存在",
	CodeVectorIndexFailed:     "向量索引构建失败",

	CodeAIServiceTimeout:         "AI 服务超时",
	CodeAIStructuredOutputFailed: "AI 结构化输出失败",
	CodeAIProviderTestFailed:     "AI Provider 连接测试失败",
	CodeAIProviderNotFound:       "AI Provider 不存在",

	CodeRateLimitExceeded: "请求过于频繁，请稍后重试",

	CodeInterviewScheduleNotFound: "面试日程不存在",
	CodeScheduleParseFailed:       "面试邀请解析失败",

	CodeVoiceSessionNotFound: "语音会话不存在",
	CodeWSConnectionFailed:   "WebSocket 连接失败",
}

func (c ErrorCode) Message() string {
	if msg, ok := defaultMessages[c]; ok {
		return msg
	}
	return "未知错误"
}

