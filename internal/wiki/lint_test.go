package wiki

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintOrphan(t *testing.T) {
	dir := t.TempDir()
	questionsDir := filepath.Join(dir, "questions")
	os.MkdirAll(questionsDir, 0755)

	// 创建题目页但不在 index 中
	os.WriteFile(filepath.Join(questionsDir, "orphan-page.md"), []byte("# 题目\n\n内容"), 0644)

	// 空 index
	idx := &Index{Nodes: make(map[string]*IndexNode)}
	idx.writeIndex(dir)

	issues, err := lint(dir, idx)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	foundOrphan := false
	for _, issue := range issues {
		if issue.Type == "orphan" && issue.Slug == "orphan-page" {
			foundOrphan = true
		}
	}
	if !foundOrphan {
		t.Error("expected orphan issue for orphan-page")
	}
}

func TestLintDanglingLink(t *testing.T) {
	dir := t.TempDir()
	questionsDir := filepath.Join(dir, "questions")
	os.MkdirAll(questionsDir, 0755)

	// 创建包含悬空链接的题目页
	os.WriteFile(filepath.Join(questionsDir, "test.md"), []byte("# 题目\n\n内容\n\n[[non-existent-page]]"), 0644)

	idx := &Index{Nodes: make(map[string]*IndexNode)}
	idx.addNode("test")
	idx.writeIndex(dir)

	issues, err := lint(dir, idx)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	foundDangling := false
	for _, issue := range issues {
		if issue.Type == "dangling_link" {
			foundDangling = true
		}
	}
	if !foundDangling {
		t.Error("expected dangling_link issue")
	}
}
