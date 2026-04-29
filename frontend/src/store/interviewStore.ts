import { create } from 'zustand';
import type {
  InterviewStage,
  Position,
  Direction,
  InterviewTurn,
  ProgrammingLanguage
} from '../types/interview';

interface InterviewState {
  interviewId: string | null;
  stage: InterviewStage;
  position: Position | null;
  direction: Direction | null;
  programmingLanguage: ProgrammingLanguage | null;
  turns: InterviewTurn[];
  isConnected: boolean;
  isProcessing: boolean;

  setInterviewId: (id: string) => void;
  setStage: (stage: InterviewStage) => void;
  setPosition: (position: Position) => void;
  setDirection: (direction: Direction) => void;
  setProgrammingLanguage: (lang: ProgrammingLanguage) => void;
  addTurn: (turn: InterviewTurn) => void;
  updateLastTurn: (updates: Partial<InterviewTurn>) => void;
  setConnected: (connected: boolean) => void;
  setProcessing: (processing: boolean) => void;
  reset: () => void;
}

const initialState = {
  interviewId: null,
  stage: 'intro' as InterviewStage,
  position: null,
  direction: null,
  programmingLanguage: null,
  turns: [],
  isConnected: false,
  isProcessing: false,
};

export const useInterviewStore = create<InterviewState>((set) => ({
  ...initialState,

  setInterviewId: (id) => set({ interviewId: id }),

  setStage: (stage) => set({ stage }),

  setPosition: (position) => set({ position }),

  setDirection: (direction) => set({ direction }),

  setProgrammingLanguage: (lang) => set({ programmingLanguage: lang }),

  addTurn: (turn) => set((state) => ({
    turns: [...state.turns, turn]
  })),

  updateLastTurn: (updates) => set((state) => {
    const turns = [...state.turns];
    if (turns.length > 0) {
      turns[turns.length - 1] = { ...turns[turns.length - 1], ...updates };
    }
    return { turns };
  }),

  setConnected: (connected) => set({ isConnected: connected }),

  setProcessing: (processing) => set({ isProcessing: processing }),

  reset: () => set(initialState),
}));
