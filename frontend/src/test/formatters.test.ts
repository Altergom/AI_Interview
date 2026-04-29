import { describe, it, expect } from 'vitest';
import {
  formatDateTime,
  formatDate,
  formatDuration,
  formatFileSize,
  formatScore,
  truncateText,
} from '../utils/formatters';

describe('formatters', () => {
  describe('formatDateTime', () => {
    it('应该格式化日期时间', () => {
      const result = formatDateTime('2026-04-29T10:30:00');
      // toLocaleString 在不同环境可能有不同格式，只检查包含关键信息
      expect(result).toContain('2026');
      expect(result).toContain('04');
      expect(result).toContain('29');
    });
  });

  describe('formatDate', () => {
    it('应该格式化日期', () => {
      const result = formatDate('2026-04-29');
      expect(result).toContain('2026');
      expect(result).toContain('04');
      expect(result).toContain('29');
    });
  });

  describe('formatDuration', () => {
    it('应该格式化秒数为分:秒', () => {
      expect(formatDuration(0)).toBe('0:00');
      expect(formatDuration(30)).toBe('0:30');
      expect(formatDuration(60)).toBe('1:00');
      expect(formatDuration(90)).toBe('1:30');
      expect(formatDuration(3661)).toBe('61:01');
    });
  });

  describe('formatFileSize', () => {
    it('应该格式化文件大小', () => {
      expect(formatFileSize(0)).toBe('0 B');
      expect(formatFileSize(1024)).toBe('1.00 KB');
      expect(formatFileSize(1024 * 1024)).toBe('1.00 MB');
      expect(formatFileSize(1024 * 1024 * 1024)).toBe('1.00 GB');
      expect(formatFileSize(1536)).toBe('1.50 KB');
    });
  });

  describe('formatScore', () => {
    it('应该格式化分数保留1位小数', () => {
      expect(formatScore(85)).toBe('85.0');
      expect(formatScore(85.5)).toBe('85.5');
      expect(formatScore(85.67)).toBe('85.7');
      expect(formatScore(85.12)).toBe('85.1');
    });
  });

  describe('truncateText', () => {
    it('应该截断长文本', () => {
      const longText = 'This is a very long text that needs to be truncated';
      expect(truncateText(longText, 20)).toBe('This is a very long ...');
    });

    it('应该保留短文本', () => {
      const shortText = 'Short text';
      expect(truncateText(shortText, 20)).toBe('Short text');
    });
  });
});
