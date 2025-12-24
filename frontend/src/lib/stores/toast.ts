import { writable } from 'svelte/store';

export type ToastLevel = 'info' | 'success' | 'warning' | 'danger';

export interface ToastAction {
  cta: string;
  callback: () => void;
}

export interface Toast {
  id: string;
  message: string;
  level: ToastLevel;
  persistent: boolean;
  dismissable: boolean;
  action?: ToastAction;
  timeout?: number;
  title?: string;
  animate?: boolean;
}

type ToastInput = Omit<Toast, 'id'>;

const DEFAULT_TIMEOUT_MS = 5000;

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([]);
  const timeouts = new Map<string, ReturnType<typeof setTimeout>>();

  function clearTimeout(id: string) {
    const existing = timeouts.get(id);
    if (existing) {
      globalThis.clearTimeout(existing);
      timeouts.delete(id);
    }
  }

  function scheduleAutoDismiss(id: string, ms: number) {
    clearTimeout(id);
    const timeout = setTimeout(() => {
      timeouts.delete(id);
      update(toasts => toasts.filter(t => t.id !== id));
    }, ms);
    timeouts.set(id, timeout);
  }

  return {
    subscribe,

    set: (id: string, toast: ToastInput) => {
      update(toasts => {
        const existing = toasts.findIndex(t => t.id === id);
        const newToast = { id, ...toast };
        if (existing >= 0) {
          toasts[existing] = newToast;
          return [...toasts];
        }
        return [...toasts, newToast];
      });

      if (toast.persistent) {
        clearTimeout(id);
      } else {
        scheduleAutoDismiss(id, toast.timeout ?? DEFAULT_TIMEOUT_MS);
      }
    },

    dismiss: (id: string) => {
      clearTimeout(id);
      update(toasts => toasts.filter(t => t.id !== id));
    },

    clear: () => {
      timeouts.forEach((_, id) => clearTimeout(id));
      update(() => []);
    }
  };
}

export const toasts = createToastStore();
