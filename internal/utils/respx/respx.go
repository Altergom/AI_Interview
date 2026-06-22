package respx

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Success(data any) Result {
	return Result{
		Code: int(CodeOK),
		Msg:  CodeOK.Message(),
		Data: data,
	}
}

func Fail(code ErrorCode) Result {
	return Result{
		Code: int(code),
		Msg:  code.Message(),
		Data: nil,
	}
}
