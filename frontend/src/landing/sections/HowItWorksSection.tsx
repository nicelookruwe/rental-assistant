import { LpCard, LpHeading, LpSection } from '../ui';
import { IconEdit, IconMonitor, IconPlus } from '../icons';

const steps = [
  {
    icon: IconPlus,
    title: 'Добавьте квартиру',
    text: 'Адрес, площадь и описание. Всё начинается с одного объекта.',
  },
  {
    icon: IconEdit,
    title: 'Заполните данные об аренде',
    text: 'Арендатор, договор, арендная плата и сроки.',
  },
  {
    icon: IconMonitor,
    title: 'Контролируйте всё из одного интерфейса',
    text: 'Платежи, сроки, счётчики и документы — всегда под рукой.',
  },
];

export function HowItWorksSection() {
  return (
    <LpSection id="lp-how">
      <LpHeading eyebrow="Как это работает">Три шага до порядка</LpHeading>
      <ol className="lp-how-grid">
        {steps.map(({ icon: Icon, title, text }, index) => (
          <li key={title}>
            <LpCard variant="outline" className="lp-step">
              <span className="lp-step-num">{index + 1}</span>
              <span className="lp-step-icon">
                <Icon />
              </span>
              <h3>{title}</h3>
              <p>{text}</p>
            </LpCard>
          </li>
        ))}
      </ol>
    </LpSection>
  );
}
