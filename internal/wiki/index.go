package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// IndexNode 表示 index.md 图中的一个节点。
type IndexNode struct {
	Slug   string
	Degree int
	Edges  []IndexEdge
}

// IndexEdge 表示节点间的一条边。
type IndexEdge struct {
	Type     string // 深挖 / 横向 / 跨域
	FromSlug string
	ToSlug   string
}

// Index 是 index.md 的内存表示。
type Index struct {
	Nodes map[string]*IndexNode // slug → node
	Edges []IndexEdge
}

// readIndex 解析 index.md 文件。
// 格式约定：
//
//	# Index
//	- [[tcp-handshake]] (3)
//	  - 深挖:[[tcp-congestion]]
//	  - 横向:[[udp-basics]]
func readIndex(wikiDir string) (*Index, error) {
	idx := &Index{Nodes: make(map[string]*IndexNode)}

	data, err := os.ReadFile(filepath.Join(wikiDir, "index.md"))
	if os.IsNotExist(err) {
		return idx, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var currentNode *IndexNode

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}

		// 节点行：- [[slug]] (N)
		if strings.HasPrefix(line, "- [[") && !strings.Contains(line, ":[[") {
			slug, degree := parseNodeLine(line)
			if slug == "" {
				continue
			}
			if existing, ok := idx.Nodes[slug]; ok {
				existing.Degree = degree
				currentNode = existing
			} else {
				currentNode = &IndexNode{Slug: slug, Degree: degree}
				idx.Nodes[slug] = currentNode
			}
			continue
		}

		// 边行：  - 深挖:[[slug]] 或   - 横向:[[slug]] 或   - 跨域:[[slug]]
		if currentNode != nil && strings.Contains(line, ":[") {
			edgeType, targetSlug := parseEdgeLine(line)
			if edgeType != "" && targetSlug != "" {
				edge := IndexEdge{Type: edgeType, FromSlug: currentNode.Slug, ToSlug: targetSlug}
				idx.Edges = append(idx.Edges, edge)
				currentNode.Edges = append(currentNode.Edges, edge)
			}
		}
	}

	return idx, nil
}

func parseNodeLine(line string) (slug string, degree int) {
	// "- [[slug]] (3)"
	line = strings.TrimPrefix(line, "- ")
	closeIdx := strings.Index(line, "]]")
	if closeIdx < 3 {
		return "", 0
	}
	slug = line[2:closeIdx]
	rest := line[closeIdx+2:]
	if n, err := fmt.Sscanf(strings.TrimSpace(rest), "(%d)", &degree); n == 1 && err == nil {
		return slug, degree
	}
	return slug, 0
}

func parseEdgeLine(line string) (edgeType, targetSlug string) {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "- ")
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		return "", ""
	}
	edgeType = line[:colonIdx]
	rest := line[colonIdx+1:]
	if strings.HasPrefix(rest, "[[") && strings.HasSuffix(rest, "]]") {
		targetSlug = rest[2 : len(rest)-2]
	}
	return edgeType, targetSlug
}

// writeIndex 将 Index 写回 index.md。
func (idx *Index) writeIndex(wikiDir string) error {
	var b strings.Builder
	b.WriteString("# Index\n\n")

	// 按 slug 排序保证稳定输出
	slugs := make([]string, 0, len(idx.Nodes))
	for slug := range idx.Nodes {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	for _, slug := range slugs {
		node := idx.Nodes[slug]
		b.WriteString(fmt.Sprintf("- [[%s]] (%d)\n", node.Slug, node.Degree))
		for _, edge := range node.Edges {
			b.WriteString(fmt.Sprintf("  - %s:[[%s]]\n", edge.Type, edge.ToSlug))
		}
		b.WriteString("\n")
	}

	return os.WriteFile(filepath.Join(wikiDir, "index.md"), []byte(b.String()), 0644)
}

// addNode 向 Index 添加一个节点（骨架，degree=0）。
func (idx *Index) addNode(slug string) {
	if _, ok := idx.Nodes[slug]; ok {
		return
	}
	idx.Nodes[slug] = &IndexNode{Slug: slug, Degree: 0}
}

// addEdges 为节点添加边并更新度数。
func (idx *Index) addEdges(fromSlug string, page QuestionPage) {
	node, ok := idx.Nodes[fromSlug]
	if !ok {
		return
	}

	for _, link := range page.FollowUp {
		if link.Slug == "" {
			continue
		}
		edge := IndexEdge{Type: link.Type, FromSlug: fromSlug, ToSlug: link.Slug}
		node.Edges = append(node.Edges, edge)
		idx.Edges = append(idx.Edges, edge)
		node.Degree++

		// 确保目标节点存在（至少作为骨架）
		if target, ok := idx.Nodes[link.Slug]; ok {
			target.Degree++
		} else {
			idx.Nodes[link.Slug] = &IndexNode{Slug: link.Slug, Degree: 1}
		}
	}
}

// removeNode 从 Index 中移除非骨架节点（用于 lint 修复后重建）。
func (idx *Index) removeNode(slug string) {
	delete(idx.Nodes, slug)
	filtered := idx.Edges[:0]
	for _, e := range idx.Edges {
		if e.FromSlug != slug && e.ToSlug != slug {
			filtered = append(filtered, e)
		} else if e.ToSlug == slug {
			if node, ok := idx.Nodes[e.FromSlug]; ok {
				node.Degree--
			}
		}
	}
	idx.Edges = filtered
}
