import { demoData } from './demoData';
import { LpBadge, LpButton, LpCard, LpHeading, LpSection, LpStat } from './ui';

export default function Styleguide() {
  return <div className="lp-root"><LpSection><LpHeading eyebrow="Rental Assistant" level={1}>Landing design system</LpHeading><p>Токены, компоненты и состояния для будущих секций.</p></LpSection><LpSection className="lp-section-tight"><LpHeading eyebrow="Buttons">Компоненты</LpHeading><div className="lp-actions"><LpButton>Primary</LpButton><LpButton variant="secondary">Secondary</LpButton><LpButton variant="ghost">Ghost</LpButton><LpButton disabled>Disabled</LpButton></div></LpSection><LpSection className="lp-section-tight"><div className="lp-grid"><LpCard><LpBadge>Primary badge</LpBadge></LpCard><LpCard variant="outline"><LpBadge tone="neutral">Outline card</LpBadge></LpCard><LpCard variant="floating-glass"><LpBadge tone="success">Glass card</LpBadge></LpCard><LpCard><LpStat value={demoData.income.value} label={demoData.income.label} detail={demoData.income.detail} /></LpCard></div></LpSection></div>;
}
