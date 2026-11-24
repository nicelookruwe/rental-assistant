import type { ButtonHTMLAttributes, ReactNode } from 'react';

type Props = ButtonHTMLAttributes<HTMLButtonElement> & { variant?: 'primary' | 'secondary' | 'ghost'; children: ReactNode };

export function LpButton({ variant = 'primary', className = '', children, ...props }: Props) {
  return <button {...props} className={`lp-button lp-button-${variant} ${className}`.trim()}>{children}</button>;
}
