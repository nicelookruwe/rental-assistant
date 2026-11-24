import type { ReactNode } from 'react';

export function LpBadge({ children, tone = 'primary' }: { children: ReactNode; tone?: 'primary' | 'neutral' | 'success' }) {
  return <span className={`lp-badge lp-badge-${tone}`}>{children}</span>;
}
