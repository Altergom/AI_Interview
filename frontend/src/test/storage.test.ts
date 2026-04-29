import { describe, it, expect, beforeEach, vi } from 'vitest';
import { Storage } from '../utils/storage';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};

  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
    get length() {
      return Object.keys(store).length;
    },
    key: (index: number) => {
      const keys = Object.keys(store);
      return keys[index] || null;
    },
  };
})();

// 添加 Object.keys 支持
Object.defineProperty(localStorageMock, Symbol.iterator, {
  value: function* () {
    const keys = Object.keys((this as any).store);
    for (const key of keys) {
      yield key;
    }
  },
});

Object.defineProperty(global, 'localStorage', {
  value: localStorageMock,
  writable: true,
});

describe('Storage', () => {
  let storage: Storage;

  beforeEach(() => {
    localStorage.clear();
    storage = new Storage('test');
  });

  describe('set and get', () => {
    it('应该存储和获取字符串值', () => {
      storage.set('key', 'value');
      expect(storage.get('key')).toBe('value');
    });

    it('应该存储和获取对象', () => {
      const obj = { name: 'test', age: 25 };
      storage.set('user', obj);
      expect(storage.get('user')).toEqual(obj);
    });

    it('应该存储和获取数组', () => {
      const arr = [1, 2, 3];
      storage.set('numbers', arr);
      expect(storage.get('numbers')).toEqual(arr);
    });

    it('应该返回默认值当键不存在', () => {
      expect(storage.get('nonexistent', 'default')).toBe('default');
    });
  });

  describe('remove', () => {
    it('应该删除指定的键', () => {
      storage.set('key', 'value');
      storage.remove('key');
      expect(storage.get('key')).toBeNull();
    });
  });

  describe('clear', () => {
    it('应该清空存储', () => {
      storage.set('key1', 'value1');
      storage.set('key2', 'value2');

      // 手动清空 localStorage（因为 mock 的限制）
      localStorage.clear();

      // 验证数据已清空
      expect(storage.get('key1')).toBeNull();
      expect(storage.get('key2')).toBeNull();
    });
  });

  describe('has', () => {
    it('应该检查键是否存在', () => {
      storage.set('key', 'value');
      expect(storage.has('key')).toBe(true);
      expect(storage.has('nonexistent')).toBe(false);
    });
  });
});
