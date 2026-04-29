import { useToastStore } from '../store/toastStore';

export const useToast = () => {
  const { showToast } = useToastStore();

  return {
    success: (message: string) => showToast(message, 'success'),
    error: (message: string) => showToast(message, 'error'),
    warning: (message: string) => showToast(message, 'warning'),
    info: (message: string) => showToast(message, 'info'),
  };
};
