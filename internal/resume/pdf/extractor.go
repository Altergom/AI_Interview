// Package pdf 提供 PDF 文本提取能力。
//
// 使用 pdfcpu 逐页读取 content stream，再解析 PDF 文字操作符（Tj / TJ / ' / "）
// 还原出可读文本。逐页处理模式避免将整份 PDF 内容一次性加载进内存。
package pdf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"

	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

const (
	// MaxPDFSize 简历 PDF 最大允许字节数（3 MB）。
	// 超出此限制的文件在读取阶段即被拒绝，防止 OOM。
	MaxPDFSize = 3 * 1024 * 1024
)

// ErrFileTooLarge 文件超过 MaxPDFSize 时返回此错误。
var ErrFileTooLarge = fmt.Errorf("pdf: file exceeds %d bytes limit", MaxPDFSize)

// ExtractText 从 r 中读取 PDF，逐页提取纯文本并拼接返回。
//
// OOM 防护：
//  1. io.LimitReader 将读取量上限锁在 MaxPDFSize+1，超出即报错
//  2. pdfcpu 内部按页处理 content stream，不会将所有页面同时展开到内存
//  3. 每页文本提取完毕立即释放 content stream bytes
func ExtractText(r io.Reader) (string, error) {
	// 1. 读取并检查大小（LimitReader 做硬限制）
	limited := io.LimitReader(r, MaxPDFSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("pdf: read: %w", err)
	}
	if len(data) > MaxPDFSize {
		return "", ErrFileTooLarge
	}
	if len(data) == 0 {
		return "", fmt.Errorf("pdf: empty file")
	}

	// 2. 用 pdfcpu 解析 PDF 结构（ReadAndValidate 不会展开所有内容流）
	rs := bytes.NewReader(data)
	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.EXTRACTCONTENT

	ctx, err := pdfapi.ReadAndValidate(rs, conf)
	if err != nil {
		return "", fmt.Errorf("pdf: parse: %w", err)
	}

	pageCount := ctx.PageCount
	if pageCount == 0 {
		return "", fmt.Errorf("pdf: no pages found")
	}

	// 3. 逐页提取 content stream → 解析文字操作符
	var sb strings.Builder
	for p := 1; p <= pageCount; p++ {
		pageReader, err := pdfcpu.ExtractPageContent(ctx, p)
		if err != nil || pageReader == nil {
			// 单页失败不中断，跳过即可
			continue
		}

		pageText, err := parseContentStream(pageReader)
		if err == nil && pageText != "" {
			sb.WriteString(pageText)
			sb.WriteByte('\n')
		}
		// pageReader 是 bytes.Reader，无需 Close
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("pdf: no extractable text found (may be image-based PDF)")
	}
	return result, nil
}

// parseContentStream 从 PDF content stream 中解析文本操作符，返回纯文本。
//
// 支持的操作符：
//   - `(string) Tj`   — 显示字符串
//   - `[(arr)] TJ`    — 显示字符串数组（忽略位移量）
//   - `(string) '`    — 换行后显示字符串
//   - `aw ac (string)"` — 换行后设字间距并显示（只取 string 部分）
//
// PDF 内容流是按行读取的文本协议，解析器为单遍扫描。
func parseContentStream(r io.Reader) (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	// 简历 PDF 单页 content stream 通常 < 512KB，给 1MB 安全余量
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	// 滑动窗口：保存当前行和上一行，用于匹配"操作数 操作符"模式
	var prev, cur string

	for scanner.Scan() {
		prev = cur
		cur = strings.TrimSpace(scanner.Text())
		if cur == "" {
			continue
		}

		switch cur {
		case "Tj":
			// 上一行应形如: (text)
			if t := extractParenString(prev); t != "" {
				sb.WriteString(t)
				sb.WriteByte(' ')
			}
		case "TJ":
			// 上一行应形如: [(text1) -100 (text2) ...]
			texts := extractTJArray(prev)
			for _, t := range texts {
				sb.WriteString(t)
			}
			if len(texts) > 0 {
				sb.WriteByte(' ')
			}
		case "'":
			// 换行 + 显示，上一行应形如: (text)
			if t := extractParenString(prev); t != "" {
				sb.WriteByte('\n')
				sb.WriteString(t)
				sb.WriteByte(' ')
			}
		case `"`:
			// aw ac (text) " —— 前几行有数字，倒数第一行是 (text)
			if t := extractParenString(prev); t != "" {
				sb.WriteByte('\n')
				sb.WriteString(t)
				sb.WriteByte(' ')
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return sb.String(), fmt.Errorf("pdf: scan content stream: %w", err)
	}
	return sb.String(), nil
}

// extractParenString 从 PDF 字符串字面量 `(text)` 中提取内容。
// 处理转义序列：\n \r \t \\ \( \)。
func extractParenString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '(' || s[len(s)-1] != ')' {
		return ""
	}
	inner := s[1 : len(s)-1]
	return unescapePDFString(inner)
}

// extractTJArray 解析 TJ 操作符的数组操作数，如 `[(Hello) -20 ( World)]`。
// 负数位移（字距调整）被忽略，只提取字符串部分。
func extractTJArray(s string) []string {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil
	}
	inner := s[1 : len(s)-1]

	var results []string
	i := 0
	for i < len(inner) {
		if inner[i] == '(' {
			// 找到配对的右括号（考虑转义）
			j := i + 1
			depth := 1
			for j < len(inner) && depth > 0 {
				if inner[j] == '\\' {
					j++ // 跳过转义字符
				} else if inner[j] == '(' {
					depth++
				} else if inner[j] == ')' {
					depth--
				}
				j++
			}
			text := unescapePDFString(inner[i+1 : j-1])
			if text != "" {
				results = append(results, text)
			}
			i = j
		} else {
			i++
		}
	}
	return results
}

// unescapePDFString 处理 PDF 字符串转义，同时过滤掉不可打印的控制字符。
func unescapePDFString(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '\\' && i+1 < len(s) {
			next := s[i+1]
			i += 2
			switch next {
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case '\\', '(', ')':
				sb.WriteByte(next)
			default:
				// 八进制转义 \nnn
				if next >= '0' && next <= '7' {
					oct := int(next - '0')
					for k := 0; k < 2 && i < len(s) && s[i] >= '0' && s[i] <= '7'; k++ {
						oct = oct*8 + int(s[i]-'0')
						i++
					}
					if oct > 0 && oct < 128 {
						sb.WriteByte(byte(oct))
					}
				}
			}
			continue
		}
		// 过滤控制字符，保留可打印 ASCII 和空白
		r := rune(c)
		if unicode.IsPrint(r) || r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			sb.WriteByte(c)
		}
		i++
	}
	return sb.String()
}
