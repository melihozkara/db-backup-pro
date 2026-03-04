import { Loader2 } from 'lucide-react';
import { cn } from '../lib/utils';

interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const sizes = {
  sm: 'w-4 h-4',
  md: 'w-6 h-6',
  lg: 'w-10 h-10',
};

export default function Spinner({ size = 'md', className = '' }: SpinnerProps) {
  return <Loader2 className={cn('animate-spin text-blue-500', sizes[size], className)} />;
}

export function PageLoader() {
  return (
    <div className="flex flex-col items-center justify-center h-[60vh] gap-3">
      <Spinner size="lg" />
      <p className="text-sm text-zinc-500">Yukleniyor...</p>
    </div>
  );
}
