import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { StructuredResume } from '../types/resume';

interface ResumeState {
  resume: StructuredResume | null;
  isParsing: boolean;
  parseError: string | null;

  setResume: (resume: StructuredResume) => void;
  updateResume: (updates: Partial<StructuredResume>) => void;
  setParsing: (parsing: boolean) => void;
  setParseError: (error: string | null) => void;
  clearResume: () => void;
}

const initialState = {
  resume: null,
  isParsing: false,
  parseError: null,
};

// 手动填表时的最小骨架——避免 resume 为 null 时丢弃 updates 导致输入框无法输入。
const emptyResume: StructuredResume = {
  user_id: '',
  skills: [],
  projects: [],
  internships: [],
  education: { school: '', major: '', graduation: '' },
};

export const useResumeStore = create<ResumeState>()(
  persist(
    (set) => ({
      ...initialState,

      setResume: (resume) => set({
        resume,
        parseError: null
      }),

      // 初次手动填表时 resume 为 null，先给一个最小骨架再合并 updates；
      // 否则 null 分支直接返回 null，更新被静默丢弃，表现为输入框无法输入。
      updateResume: (updates) => set((state) => ({
        resume: {
          ...emptyResume,
          ...(state.resume ?? {}),
          ...updates,
        },
      })),

      setParsing: (parsing) => set({ isParsing: parsing }),

      setParseError: (error) => set({
        parseError: error,
        isParsing: false
      }),

      clearResume: () => set(initialState),
    }),
    {
      name: 'resume-storage',
    }
  )
);
