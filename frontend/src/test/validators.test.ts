import { describe, it, expect } from 'vitest';
import {
  validateEmail,
  validatePassword,
  validateUsername,
  validateRequired,
  validateFileType,
  validateFileSize,
} from '../utils/validators';

describe('validators', () => {
  describe('validateEmail', () => {
    it('应该验证有效的邮箱地址', () => {
      expect(validateEmail('test@example.com')).toBe(true);
      expect(validateEmail('user.name@domain.co.uk')).toBe(true);
      expect(validateEmail('user+tag@example.com')).toBe(true);
    });

    it('应该拒绝无效的邮箱地址', () => {
      expect(validateEmail('invalid')).toBe(false);
      expect(validateEmail('test@')).toBe(false);
      expect(validateEmail('@example.com')).toBe(false);
      expect(validateEmail('test @example.com')).toBe(false);
      expect(validateEmail('')).toBe(false);
    });
  });

  describe('validatePassword', () => {
    it('应该验证有效的密码', () => {
      expect(validatePassword('Password123')).toBe(true);
      expect(validatePassword('Test1234')).toBe(true);
      expect(validatePassword('abcd1234')).toBe(true);
    });

    it('应该拒绝无效的密码', () => {
      expect(validatePassword('short1')).toBe(false); // 太短
      expect(validatePassword('NoNumbers')).toBe(false); // 没有数字
      expect(validatePassword('12345678')).toBe(false); // 没有字母
      expect(validatePassword('')).toBe(false); // 空字符串
    });
  });

  describe('validateUsername', () => {
    it('应该验证有效的用户名', () => {
      expect(validateUsername('user')).toBe(true);
      expect(validateUsername('test_user')).toBe(true);
      expect(validateUsername('user123')).toBe(true);
    });

    it('应该拒绝无效的用户名', () => {
      expect(validateUsername('a')).toBe(false); // 太短
      expect(validateUsername('a'.repeat(21))).toBe(false); // 太长
      expect(validateUsername('')).toBe(false); // 空字符串
    });
  });

  describe('validateRequired', () => {
    it('应该验证非空值', () => {
      expect(validateRequired('value')).toBe(true);
      expect(validateRequired('0')).toBe(true);
    });

    it('应该拒绝空值', () => {
      expect(validateRequired('')).toBe(false);
      expect(validateRequired('   ')).toBe(false);
    });
  });

  describe('validateFileType', () => {
    it('应该验证正确的文件类型', () => {
      const pdfFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      expect(validateFileType(pdfFile, ['application/pdf'])).toBe(true);

      const docFile = new File(['content'], 'test.doc', { type: 'application/msword' });
      expect(validateFileType(docFile, ['application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'])).toBe(true);
    });

    it('应该拒绝错误的文件类型', () => {
      const txtFile = new File(['content'], 'test.txt', { type: 'text/plain' });
      expect(validateFileType(txtFile, ['application/pdf'])).toBe(false);
    });
  });

  describe('validateFileSize', () => {
    it('应该验证文件大小在限制内', () => {
      const smallFile = new File(['a'.repeat(100)], 'small.txt');
      expect(validateFileSize(smallFile, 1024)).toBe(true);
    });

    it('应该拒绝超过大小限制的文件', () => {
      const largeFile = new File(['a'.repeat(2000)], 'large.txt');
      expect(validateFileSize(largeFile, 1024)).toBe(false);
    });
  });
});
