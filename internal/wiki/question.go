package wiki

import (
	"fmt"
	"strings"
)

// QuestionPage 是 LLM 生成的结构化题目页。
type QuestionPage struct {
	Slug         string         `json:"slug"`
	Question     string         `json:"question"`
	FocusPoints  string         `json:"focus_points"`
	AnswerPoints []string       `json:"answer_points"`
	FollowUp     []FollowUpLink `json:"follow_up"`
	Tags         []string       `json:"tags"`
	Difficulty   string         `json:"difficulty"`
}

// FollowUpLink 追问链接。
type FollowUpLink struct {
	Type  string `json:"type"`  // 深挖 / 横向 / 跨域
	Slug  string `json:"slug"`
	Label string `json:"label"`
}

// renderMarkdown 按 schema.md 模板渲染题目页 Markdown。
func (p *QuestionPage) renderMarkdown(source string) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(p.Tags, ", ")))
	b.WriteString(fmt.Sprintf("source: raw/%s\n", source))
	b.WriteString("---\n\n")

	b.WriteString("# 题目\n\n")
	b.WriteString(p.Question)
	b.WriteString("\n\n")

	b.WriteString("# 考察点\n\n")
	b.WriteString(p.FocusPoints)
	b.WriteString("\n\n")

	b.WriteString("# 答案要点\n\n")
	for _, pt := range p.AnswerPoints {
		b.WriteString(fmt.Sprintf("- %s\n", pt))
	}
	b.WriteString("\n")

	b.WriteString("# 追问\n\n")
	for _, link := range p.FollowUp {
		b.WriteString(fmt.Sprintf("%s: [[%s]]\n", link.Type, link.Slug))
	}

	return b.String()
}
