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

export const useResumeStore = create<ResumeState>()(
  persist(
    (set) => ({
      ...initialState,

      setResume: (resume) => set({
        resume,
        parseError: null
      }),

      updateResume: (updates) => set((state) => ({
        resume: state.resume ? { ...state.resume, ...updates } : null
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
