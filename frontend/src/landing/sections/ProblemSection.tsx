import { LpCard, LpHeading, LpSection } from '../ui';
import { IconCalendar, IconDoc, IconTable, IconWallet } from '../icons';

const problems = [
  {
    icon: IconDoc,
    title: 'Разрозненные документы',
    text: 'Договоры, акты и чеки лежат в папках, чатах и почте.',
  },
  {
    icon: IconWallet,
    title: 'Забытые платежи',
    text: 'Кто и за что заплатил — приходится вспоминать и перепроверять.',
  },
  {
    icon: IconCalendar,
    title: 'Неизвестные сроки договоров',
    text: 'Окончание договора застаёт врасплох, продление теряется.',
  },
  {
    icon: IconTable,
    title: 'Таблицы с доходами и расходами',
    text: 'Сводки живут в таблицах, которые легко забыть обновить.',
  },
];

export function ProblemSection() {
  return (
    <LpSection id="lp-problem">
      <LpHeading eyebrow="Проблема">
        Таблицы, заметки и переписки — не система
      </LpHeading>
      <div className="lp-grid">
        {problems.map(({ icon: Icon, title, text }) => (
          <LpCard key={title} variant="outline" className="lp-problem-card">
            <span className="lp-problem-icon">
              <Icon />
            </span>
            <h3>{title}</h3>
            <p>{text}</p>
          </LpCard>
        ))}
      </div>
      <p className="lp-problem-cta">Пора собрать всё в одном месте.</p>
    </LpSection>
  );
}
