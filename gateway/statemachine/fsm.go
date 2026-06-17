package statemachine

import (
	"context"
	"fmt"

	"gateway/domain"
)

// Action 状态机决定网关下一步做什么。
type Action string

const (
	ActionTriggerAgent  Action = "trigger_agent"  // 调用 Agent 处理本轮消息
	ActionSendMessage   Action = "send_message"    // 直接回复固定文本（onboarding 引导等）
	ActionPause         Action = "pause"           // 暂停面试
	ActionHandoff       Action = "handoff"         // 转人工
	ActionEnd           Action = "end"             // 结束面试
	ActionReject        Action = "reject"          // 拒绝本条消息（状态不允许）
	ActionNoop          Action = "noop"            // 无操作
)

// Decision 状态机决策结果。
type Decision struct {
	Action  Action
	Message string // ActionSendMessage 时填充回复文本
}

// FSM 面试网关状态机。
// 决定每条消息进来后网关如何推进，最终执行权始终属于网关而非 Agent。
type FSM struct{}

func NewFSM() *FSM { return &FSM{} }

// Decide 根据当前会话状态和入站事件类型，返回网关应执行的动作。
func (f *FSM) Decide(ctx context.Context, session *domain.GatewaySession, eventType string) (*Decision, error) {
	switch session.Status {
	case domain.StatusNew, domain.StatusVerifying:
		return &Decision{Action: ActionSendMessage, Message: "请先完成身份绑定"}, nil

	case domain.StatusReady:
		return &Decision{Action: ActionTriggerAgent}, nil

	case domain.StatusInterviewing:
		return &Decision{Action: ActionTriggerAgent}, nil

	case domain.StatusWaitingUser:
		return &Decision{Action: ActionTriggerAgent}, nil

	case domain.StatusPaused:
		return &Decision{Action: ActionReject, Message: "面试已暂停，请稍后再试"}, nil

	case domain.StatusFinished, domain.StatusExpired:
		return &Decision{Action: ActionReject, Message: "面试已结束"}, nil

	case domain.StatusHandoff:
		return &Decision{Action: ActionNoop}, nil

	default:
		return nil, fmt.Errorf("unknown session status: %s", session.Status)
	}
}

// Transition 执行状态迁移，校验迁移是否合法。
func (f *FSM) Transition(ctx context.Context, session *domain.GatewaySession, target domain.SessionStatus) error {
	// TODO: 定义合法迁移表并校验
	return nil
}
