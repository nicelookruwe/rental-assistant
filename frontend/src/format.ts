const moneyFormatter = new Intl.NumberFormat('ru-RU', {
  style: 'currency',
  currency: 'RUB',
  minimumFractionDigits: 0,
  maximumFractionDigits: 2,
});

const numberFormatter = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 2 });

export function money(value: number | undefined | null): string {
  return moneyFormatter.format(value ?? 0);
}

export function num(value: number | undefined | null): string {
  return numberFormatter.format(value ?? 0);
}

export function date(value?: string | null): string {
  if (!value) return '—';
  const d = new Date(value.length <= 10 ? `${value}T00:00:00` : value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
}

export function dateTime(value?: string | null): string {
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function fileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} Б`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} КБ`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} МБ`;
}

export function today(): string {
  return new Date().toISOString().slice(0, 10);
}

export function monthLabel(month: string): string {
  const [year, m] = month.split('-');
  const d = new Date(Number(year), Number(m) - 1, 1);
  return d.toLocaleDateString('ru-RU', { month: 'short' });
}
