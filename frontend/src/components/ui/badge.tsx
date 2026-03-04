import * as React from 'react';
import { cn } from '../../lib/utils';

const variants = {
  default: 'bg-zinc-800 text-zinc-300 border-zinc-700',
  success: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30',
  destructive: 'bg-red-500/15 text-red-400 border-red-500/30',
  warning: 'bg-amber-500/15 text-amber-400 border-amber-500/30',
  info: 'bg-blue-500/15 text-blue-400 border-blue-500/30',
  purple: 'bg-purple-500/15 text-purple-400 border-purple-500/30',
  orange: 'bg-orange-500/15 text-orange-400 border-orange-500/30',
};

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: keyof typeof variants;
}

function Badge({ className, variant = 'default', ...props }: BadgeProps) {
  return (
    <div
      className={cn(
        'inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium transition-colors',
        variants[variant],
        className
      )}
      {...props}
    />
  );
}

export { Badge };
