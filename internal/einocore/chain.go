package einocore

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
)

// NewPassthroughChain 占位：string -> string 恒等链，用于验证 Eino compose 依赖与编译。
// 后续 Interview / Code Judge 等 Graph 可替换为真实节点编排（见 eino-guide）。
func NewPassthroughChain(ctx context.Context) (compose.Runnable[string, string], error) {
	ch := compose.NewChain[string, string]()
	ch.AppendLambda(compose.InvokableLambda(func(ctx context.Context, in string) (string, error) {
		_ = ctx
		return in, nil
	}))
	r, err := ch.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile passthrough chain: %w", err)
	}
	return r, nil
}
