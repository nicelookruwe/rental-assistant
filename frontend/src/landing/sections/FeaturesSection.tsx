import { LpBadge, LpCard, LpHeading, LpSection } from '../ui';
import { IconChart, IconDoc, IconGauge, IconHome, IconWallet } from '../icons';

const features = [
  {
    icon: IconHome,
    title: 'Квартиры',
    text: 'Все объекты в одном списке: адрес, площадь и действующий договор. '
      + 'Открываете — и видите картину целиком.',
  },
  {
    icon: IconDoc,
    title: 'Договоры аренды',
    text: 'Сроки, арендная плата и депозит. Напоминания не дают пропустить '
      + 'окончание договора.',
  },
  {
    icon: IconWallet,
    title: 'Учёт платежей',
    text: 'История арендных и коммунальных платежей. Видно, что оплачено, '
      + 'а что ещё ожидает оплаты.',
  },
  {
    icon: IconChart,
    title: 'Доходы и расходы',
    text: 'Сводка по месяцам: сколько пришло, сколько ушло и что осталось.',
  },
  {
    icon: IconGauge,
    title: 'Счётчики и документы',
    text: 'Показания без потери истории и документы с фото по каждой квартире.',
  },
];

export function FeaturesSection() {
  return (
    <LpSection id="lp-features">
      <LpHeading eyebrow="Возможности">
        Всё, что нужно для порядка в аренде
      </LpHeading>
      <div className="lp-feature-grid">
        {features.map(({ icon: Icon, title, text }) => (
          <LpCard key={title} className="lp-feature-card">
            <span className="lp-feature-icon">
              <Icon />
            </span>
            <h3>{title}</h3>
            <p>{text}</p>
          </LpCard>
        ))}
      </div>
      <div className="lp-audience">
        <LpBadge>Для кого</LpBadge>
        <h3>Для тех, у кого 1–3 квартиры</h3>
        <p>
          Без лишних функций — только договоры, платежи, счётчики
          и документы в одном месте.
        </p>
      </div>
    </LpSection>
  );
}
