package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LintIssue 表示一个 lint 问题。
type LintIssue struct {
	Type    string // orphan / dangling_link
	Slug    string
	Detail  string
}

// lint 按 schema.md 的 Lint 规则检查 wiki 目录。
func lint(wikiDir string, idx *Index) ([]LintIssue, error) {
	var issues []LintIssue

	// 构建 slug → 是否存在题目页的集合
	questionSlugs := make(map[string]bool)
	entries, err := os.ReadDir(filepath.Join(wikiDir, "questions"))
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			slug := strings.TrimSuffix(e.Name(), ".md")
			questionSlugs[slug] = true
		}
	}

	// 孤立节点：questions/ 下存在但在 index.md 中度数为 0
	for slug := range questionSlugs {
		node, ok := idx.Nodes[slug]
		if !ok {
			issues = append(issues, LintIssue{
				Type:   "orphan",
				Slug:   slug,
				Detail: "questions/ 下存在但 index.md 中无此节点",
			})
		} else if node.Degree == 0 {
			issues = append(issues, LintIssue{
				Type:   "orphan",
				Slug:   slug,
				Detail: "度数为 0，需人工连线",
			})
		}
	}

	// 悬空链接：题目页中的 wikilink 指向不存在的文件
	wikiLinkRe := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	for slug := range questionSlugs {
		content, err := os.ReadFile(filepath.Join(wikiDir, "questions", slug+".md"))
		if err != nil {
			continue
		}
		matches := wikiLinkRe.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			targetSlug := m[1]
			if !questionSlugs[targetSlug] {
				issues = append(issues, LintIssue{
					Type:   "dangling_link",
					Slug:   slug,
					Detail: fmt.Sprintf("悬空链接 [[%s]]", targetSlug),
				})
			}
		}
	}

	return issues, nil
}
