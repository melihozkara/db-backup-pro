import { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import {
  AlertDialog, AlertDialogContent, AlertDialogHeader,
  AlertDialogTitle, AlertDialogDescription, AlertDialogFooter,
  AlertDialogAction, AlertDialogCancel,
} from './ui/alert-dialog';
import { AlertTriangle, AlertCircle, Info } from 'lucide-react';

interface ConfirmOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  type?: 'danger' | 'warning' | 'info';
}

interface ConfirmContextType {
  confirm: (options: ConfirmOptions) => Promise<boolean>;
}

const ConfirmContext = createContext<ConfirmContextType | null>(null);

export function useConfirm() {
  const context = useContext(ConfirmContext);
  if (!context) {
    throw new Error('useConfirm must be used within ConfirmProvider');
  }
  return context;
}

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [dialog, setDialog] = useState<(ConfirmOptions & { resolve: (value: boolean) => void }) | null>(null);

  const confirm = useCallback((options: ConfirmOptions): Promise<boolean> => {
    return new Promise(resolve => {
      setDialog({ ...options, resolve });
    });
  }, []);

  const handleConfirm = () => {
    dialog?.resolve(true);
    setDialog(null);
  };

  const handleCancel = () => {
    dialog?.resolve(false);
    setDialog(null);
  };

  const getIcon = (type: string = 'danger') => {
    const iconClass = 'w-5 h-5';
    switch (type) {
      case 'danger':
        return (
          <div className="w-10 h-10 rounded-full bg-red-500/15 flex items-center justify-center">
            <AlertTriangle className={`${iconClass} text-red-400`} />
          </div>
        );
      case 'warning':
        return (
          <div className="w-10 h-10 rounded-full bg-amber-500/15 flex items-center justify-center">
            <AlertCircle className={`${iconClass} text-amber-400`} />
          </div>
        );
      case 'info':
        return (
          <div className="w-10 h-10 rounded-full bg-blue-500/15 flex items-center justify-center">
            <Info className={`${iconClass} text-blue-400`} />
          </div>
        );
    }
  };

  return (
    <ConfirmContext.Provider value={{ confirm }}>
      {children}
      <AlertDialog open={!!dialog} onOpenChange={(open) => { if (!open) handleCancel(); }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <div className="flex items-start gap-4">
              {getIcon(dialog?.type)}
              <div className="flex-1">
                <AlertDialogTitle>{dialog?.title}</AlertDialogTitle>
                <AlertDialogDescription className="mt-2">{dialog?.message}</AlertDialogDescription>
              </div>
            </div>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleCancel}>
              {dialog?.cancelText || 'Iptal'}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirm}
              className={dialog?.type === 'danger' ? 'bg-red-600 hover:bg-red-700' : 'bg-blue-600 hover:bg-blue-700'}
            >
              {dialog?.confirmText || 'Onayla'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </ConfirmContext.Provider>
  );
}
