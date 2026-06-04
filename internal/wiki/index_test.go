package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadWriteIndex(t *testing.T) {
	dir := t.TempDir()

	// 初始状态下无 index.md，readIndex 返回空 Index
	idx, err := readIndex(dir)
	if err != nil {
		t.Fatalf("readIndex: %v", err)
	}
	if len(idx.Nodes) != 0 {
		t.Fatalf("expected empty index, got %d nodes", len(idx.Nodes))
	}

	// 添加节点和边
	idx.addNode("tcp-handshake")
	idx.addEdges("tcp-handshake", QuestionPage{
		FollowUp: []FollowUpLink{
			{Type: "深挖", Slug: "tcp-congestion", Label: "TCP 拥塞控制"},
			{Type: "横向", Slug: "udp-basics", Label: "UDP 基础"},
		},
	})

	if err := idx.writeIndex(dir); err != nil {
		t.Fatalf("writeIndex: %v", err)
	}

	// 重新读取并验证
	idx2, err := readIndex(dir)
	if err != nil {
		t.Fatalf("readIndex after write: %v", err)
	}

	node, ok := idx2.Nodes["tcp-handshake"]
	if !ok {
		t.Fatal("expected tcp-handshake node")
	}
	if node.Degree != 2 {
		t.Fatalf("expected degree 2, got %d", node.Degree)
	}

	target, ok := idx2.Nodes["tcp-congestion"]
	if !ok {
		t.Fatal("expected tcp-congestion node")
	}
	if target.Degree != 1 {
		t.Fatalf("expected degree 1 for target, got %d", target.Degree)
	}
}

func TestParseNodeLine(t *testing.T) {
	slug, degree := parseNodeLine("- [[tcp-handshake]] (3)")
	if slug != "tcp-handshake" || degree != 3 {
		t.Fatalf("expected tcp-handshake/3, got %s/%d", slug, degree)
	}
}

func TestParseEdgeLine(t *testing.T) {
	edgeType, targetSlug := parseEdgeLine("  - 深挖:[[tcp-congestion]]")
	if edgeType != "深挖" || targetSlug != "tcp-congestion" {
		t.Fatalf("expected 深挖/tcp-congestion, got %s/%s", edgeType, targetSlug)
	}
}

func TestAppendLog(t *testing.T) {
	dir := t.TempDir()

	if err := appendLog(dir, "CS-Base/网络/tcp.md", "tcp-handshake"); err != nil {
		t.Fatalf("appendLog: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "log.md"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	if !strings.Contains(string(data), "ingest") {
		t.Error("expected ingest entry")
	}
	if !strings.Contains(string(data), "tcp-handshake") {
		t.Error("expected slug in log")
	}
}
