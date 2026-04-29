import { create } from 'zustand';

interface DeviceState {
  microphonePermission: PermissionState | null;
  cameraPermission: PermissionState | null;
  microphoneDeviceId: string | null;
  cameraDeviceId: string | null;
  microphoneStream: MediaStream | null;
  cameraStream: MediaStream | null;
  isMicrophoneTested: boolean;
  isCameraTested: boolean;
  audioLevel: number;

  setMicrophonePermission: (state: PermissionState) => void;
  setCameraPermission: (state: PermissionState) => void;
  setMicrophoneDeviceId: (id: string) => void;
  setCameraDeviceId: (id: string) => void;
  setMicrophoneStream: (stream: MediaStream | null) => void;
  setCameraStream: (stream: MediaStream | null) => void;
  setMicrophoneTested: (tested: boolean) => void;
  setCameraTested: (tested: boolean) => void;
  setAudioLevel: (level: number) => void;
  stopAllStreams: () => void;
  reset: () => void;
}

const initialState = {
  microphonePermission: null,
  cameraPermission: null,
  microphoneDeviceId: null,
  cameraDeviceId: null,
  microphoneStream: null,
  cameraStream: null,
  isMicrophoneTested: false,
  isCameraTested: false,
  audioLevel: 0,
};

export const useDeviceStore = create<DeviceState>((set, get) => ({
  ...initialState,

  setMicrophonePermission: (state) => set({ microphonePermission: state }),

  setCameraPermission: (state) => set({ cameraPermission: state }),

  setMicrophoneDeviceId: (id) => set({ microphoneDeviceId: id }),

  setCameraDeviceId: (id) => set({ cameraDeviceId: id }),

  setMicrophoneStream: (stream) => set({ microphoneStream: stream }),

  setCameraStream: (stream) => set({ cameraStream: stream }),

  setMicrophoneTested: (tested) => set({ isMicrophoneTested: tested }),

  setCameraTested: (tested) => set({ isCameraTested: tested }),

  setAudioLevel: (level) => set({ audioLevel: level }),

  stopAllStreams: () => {
    const { microphoneStream, cameraStream } = get();

    if (microphoneStream) {
      microphoneStream.getTracks().forEach(track => track.stop());
    }

    if (cameraStream) {
      cameraStream.getTracks().forEach(track => track.stop());
    }

    set({
      microphoneStream: null,
      cameraStream: null
    });
  },

  reset: () => {
    get().stopAllStreams();
    set(initialState);
  },
}));
