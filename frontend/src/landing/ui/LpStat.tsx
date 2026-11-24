import type { ReactNode } from 'react';

export function LpStat({ value, label, detail }: { value: ReactNode; label: string; detail?: ReactNode }) {
  return <div className="lp-stat"><strong>{value}</strong><span>{label}</span>{detail ? <small>{detail}</small> : null}</div>;
}
