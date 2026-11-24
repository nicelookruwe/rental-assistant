import type { ReactNode } from 'react';

export function LpHeading({ eyebrow, children, level = 2 }: { eyebrow?: string; children: ReactNode; level?: 1 | 2 | 3 }) {
  const Heading = `h${level}` as 'h1' | 'h2' | 'h3';
  return <div className="lp-heading">{eyebrow ? <span className="lp-eyebrow">{eyebrow}</span> : null}<Heading>{children}</Heading></div>;
}
