package wiki

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	page := &QuestionPage{
		Slug:        "tcp-handshake",
		Question:    "TCP 三次握手的过程是什么？",
		FocusPoints: "TCP 连接建立的流程和状态转换",
		AnswerPoints: []string{
			"第一次：客户端发送 SYN",
			"第二次：服务端回复 SYN-ACK",
			"第三次：客户端发送 ACK",
		},
		FollowUp: []FollowUpLink{
			{Type: "深挖", Slug: "tcp-congestion", Label: "TCP 拥塞控制"},
			{Type: "横向", Slug: "udp-basics", Label: "UDP 基础"},
		},
		Tags:       []string{"#计算机网络", "#难度:中"},
		Difficulty: "中",
	}

	result := page.renderMarkdown("raw/CS-Base/网络/tcp-handshake.md")

	if !strings.Contains(result, "source: raw/raw/") && !strings.Contains(result, "source: raw/CS-Base") {
		t.Error("expected source in frontmatter")
	}
	if !strings.Contains(result, "# 题目") {
		t.Error("expected 题目 section")
	}
	if !strings.Contains(result, "# 考察点") {
		t.Error("expected 考察点 section")
	}
	if !strings.Contains(result, "# 答案要点") {
		t.Error("expected 答案要点 section")
	}
	if !strings.Contains(result, "# 追问") {
		t.Error("expected 追问 section")
	}
	if !strings.Contains(result, "深挖: [[tcp-congestion]]") {
		t.Error("expected 深挖 link")
	}
	if !strings.Contains(result, "横向: [[udp-basics]]") {
		t.Error("expected 横向 link")
	}
}
