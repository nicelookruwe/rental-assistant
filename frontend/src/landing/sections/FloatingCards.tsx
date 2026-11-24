import { demoData } from '../demoData';
import { LpBadge, LpCard, LpStat } from '../ui';

export function FloatingCards() {
  return <div className="lp-floating-cards" aria-label="Ключевые показатели"><LpCard variant="floating-glass" className="lp-float-card lp-float-payment"><LpBadge tone="success">Получено</LpBadge><LpStat value={demoData.payment.value} label={demoData.payment.label} detail={demoData.payment.detail} /></LpCard><LpCard variant="floating-glass" className="lp-float-card lp-float-contract"><LpBadge tone="neutral">Внимание</LpBadge><LpStat value="11 дней" label={demoData.contract.label} detail={demoData.contract.detail} /></LpCard><LpCard variant="floating-glass" className="lp-float-card lp-float-income"><LpStat value={demoData.income.value} label={demoData.income.label} detail={demoData.income.detail} /></LpCard><LpCard variant="floating-glass" className="lp-float-card lp-float-meter"><LpStat value={demoData.meter.value} label={demoData.meter.label} detail={demoData.meter.detail} /></LpCard></div>;
}
