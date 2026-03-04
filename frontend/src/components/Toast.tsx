import { createContext, useContext, useCallback, ReactNode } from 'react';
import { Toaster, toast as sonnerToast } from 'sonner';

type ToastType = 'success' | 'error' | 'warning' | 'info';

interface ToastContextType {
  showToast: (type: ToastType, title: string, message?: string) => void;
  success: (title: string, message?: string) => void;
  error: (title: string, message?: string) => void;
  warning: (title: string, message?: string) => void;
  info: (title: string, message?: string) => void;
}

const ToastContext = createContext<ToastContextType | null>(null);

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const showToast = useCallback((type: ToastType, title: string, message?: string) => {
    const options = { description: message, duration: 4000 };
    switch (type) {
      case 'success': sonnerToast.success(title, options); break;
      case 'error': sonnerToast.error(title, options); break;
      case 'warning': sonnerToast.warning(title, options); break;
      case 'info': sonnerToast.info(title, options); break;
    }
  }, []);

  const success = useCallback((title: string, message?: string) => showToast('success', title, message), [showToast]);
  const error = useCallback((title: string, message?: string) => showToast('error', title, message), [showToast]);
  const warning = useCallback((title: string, message?: string) => showToast('warning', title, message), [showToast]);
  const info = useCallback((title: string, message?: string) => showToast('info', title, message), [showToast]);

  return (
    <ToastContext.Provider value={{ showToast, success, error, warning, info }}>
      {children}
      <Toaster
        position="bottom-right"
        theme="dark"
        toastOptions={{
          style: {
            background: '#18181b',
            border: '1px solid #27272a',
            color: '#fafafa',
          },
        }}
        richColors
        closeButton
      />
    </ToastContext.Provider>
  );
}
