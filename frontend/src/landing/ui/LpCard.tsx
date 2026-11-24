import type { ReactNode } from 'react';

export function LpCard({ variant = 'surface', className = '', children }: { variant?: 'surface' | 'outline' | 'floating-glass'; className?: string; children: ReactNode }) {
  return <div className={`lp-card lp-card-${variant} ${className}`.trim()}>{children}</div>;
}
