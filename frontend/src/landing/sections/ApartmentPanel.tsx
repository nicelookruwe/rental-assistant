import { useState } from 'react';
import { demoData } from '../demoData';
import { LpBadge, LpCard } from '../ui';

const apartments = ['A-01', 'A-02', 'B-01', 'B-02', 'C-01', 'C-02', 'D-01', 'D-02'];

export function ApartmentPanel() {
  const [selected, setSelected] = useState(apartments[0]);
  return <div className="lp-apartment-panel"><div className="lp-apartment-list" role="list" aria-label="Выбрать квартиру">{apartments.map((apartment) => <button type="button" key={apartment} className={`lp-apartment-button${selected === apartment ? ' is-selected' : ''}`} aria-label={`Выбрать квартиру ${apartment}`} aria-pressed={selected === apartment} onClick={() => setSelected(apartment)}>{apartment}</button>)}</div><LpCard variant="floating-glass" className="lp-selected-apartment"><div className="lp-selected-head"><div><span className="lp-eyebrow">Квартира {selected}</span><strong>Анна Петрова</strong></div><LpBadge tone="success">Оплата получена</LpBadge></div><dl className="lp-detail-list"><div><dt>Договор</dt><dd>{demoData.contract.detail}</dd></div><div><dt>Статус платежа</dt><dd>{demoData.payment.detail}</dd></div></dl></LpCard></div>;
}
