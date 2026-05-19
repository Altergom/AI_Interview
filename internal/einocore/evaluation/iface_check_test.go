package evaluation

import "testing"

// 编译期断言：Evaluator 实现 EvaluationPipeline 接口。
var _ EvaluationPipeline = (*Evaluator)(nil)

func TestIfaceCheck(t *testing.T) {}
