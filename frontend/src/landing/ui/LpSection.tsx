import type { ReactNode } from 'react';

export function LpSection({ children, className = '', id }: { children: ReactNode; className?: string; id?: string }) {
  return <section id={id} className={`lp-section ${className}`.trim()}><div className="lp-container">{children}</div></section>;
}
