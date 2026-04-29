import { STORAGE_KEYS } from './constants';

// 通用 localStorage 操作封装
export class Storage {
  private prefix: string;

  constructor(prefix: string = '') {
    this.prefix = prefix;
  }

  private getKey(key: string): string {
    return this.prefix ? `${this.prefix}:${key}` : key;
  }

  // 保存数据
  set<T>(key: string, value: T): void {
    try {
      const serialized = JSON.stringify(value);
      localStorage.setItem(this.getKey(key), serialized);
    } catch (error) {
      console.error(`Failed to save to localStorage: ${key}`, error);
    }
  }

  // 获取数据
  get<T>(key: string, defaultValue: T | null = null): T | null {
    try {
      const item = localStorage.getItem(this.getKey(key));
      if (item === null) return defaultValue;
      return JSON.parse(item) as T;
    } catch (error) {
      console.error(`Failed to read from localStorage: ${key}`, error);
      return defaultValue;
    }
  }

  // 删除数据
  remove(key: string): void {
    try {
      localStorage.removeItem(this.getKey(key));
    } catch (error) {
      console.error(`Failed to remove from localStorage: ${key}`, error);
    }
  }

  // 清空所有带前缀的数据
  clear(): void {
    try {
      if (this.prefix) {
        const keys = Object.keys(localStorage);
        keys.forEach(key => {
          if (key.startsWith(`${this.prefix}:`)) {
            localStorage.removeItem(key);
          }
        });
      } else {
        localStorage.clear();
      }
    } catch (error) {
      console.error('Failed to clear localStorage', error);
    }
  }

  // 检查键是否存在
  has(key: string): boolean {
    return localStorage.getItem(this.getKey(key)) !== null;
  }
}

const storage = new Storage();

// 认证相关存储
export const authStorage = {
  setToken: (token: string) => storage.set(STORAGE_KEYS.TOKEN, token),
  getToken: () => storage.get<string>(STORAGE_KEYS.TOKEN),
  removeToken: () => storage.remove(STORAGE_KEYS.TOKEN),

  setUserId: (userId: string) => storage.set(STORAGE_KEYS.USER_ID, userId),
  getUserId: () => storage.get<string>(STORAGE_KEYS.USER_ID),
  removeUserId: () => storage.remove(STORAGE_KEYS.USER_ID),

  setUsername: (username: string) => storage.set(STORAGE_KEYS.USERNAME, username),
  getUsername: () => storage.get<string>(STORAGE_KEYS.USERNAME),
  removeUsername: () => storage.remove(STORAGE_KEYS.USERNAME),

  clearAuth: () => {
    storage.remove(STORAGE_KEYS.TOKEN);
    storage.remove(STORAGE_KEYS.USER_ID);
    storage.remove(STORAGE_KEYS.USERNAME);
  },
};

// 面试相关存储
export const interviewStorage = {
  setInterviewId: (interviewId: string) =>
    storage.set(STORAGE_KEYS.INTERVIEW_ID, interviewId),
  getInterviewId: () => storage.get<string>(STORAGE_KEYS.INTERVIEW_ID),
  removeInterviewId: () => storage.remove(STORAGE_KEYS.INTERVIEW_ID),
};

// 简历草稿存储
export const resumeStorage = {
  saveDraft: (draft: any) => storage.set(STORAGE_KEYS.RESUME_DRAFT, draft),
  getDraft: () => storage.get<any>(STORAGE_KEYS.RESUME_DRAFT),
  clearDraft: () => storage.remove(STORAGE_KEYS.RESUME_DRAFT),
};

export default storage;
