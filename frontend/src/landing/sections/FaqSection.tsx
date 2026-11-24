import { useRef, useState } from 'react';
import type { KeyboardEvent } from 'react';
import { LpHeading, LpSection } from '../ui';
import { IconChevron } from '../icons';

const faqs = [
  {
    q: 'Кому подойдёт сервис?',
    a: 'Тем, кто сдаёт одну–три квартиры и хочет держать договоры, платежи '
      + 'и документы в одном месте.',
  },
  {
    q: 'Что нужно ввести в начале?',
    a: 'Адрес квартиры, данные арендатора и условия договора. Остальное '
      + 'добавляется по ходу.',
  },
  {
    q: 'Кто видит мои данные?',
    a: 'Данные привязаны к вашему кабинету и доступны только после входа в него.',
  },
  {
    q: 'Что происходит со сроком договора?',
    a: 'Напоминание приходит заранее, чтобы вы успели подготовиться к продлению.',
  },
  {
    q: 'Как учитываются платежи?',
    a: 'Каждое начисление и оплата сохраняются отдельными записями, история '
      + 'не перезаписывается.',
  },
  {
    q: 'Что с показаниями счётчиков?',
    a: 'Каждое показание — отдельная запись. Предыдущие значения сохраняются.',
  },
];

export function FaqSection() {
  const [openIndex, setOpenIndex] = useState<number | null>(0);
  const buttonsRef = useRef<Array<HTMLButtonElement | null>>([]);

  function toggle(index: number) {
    setOpenIndex((current) => (current === index ? null : index));
  }

  function onKeyDown(event: KeyboardEvent<HTMLButtonElement>, index: number) {
    const buttons = buttonsRef.current;
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      const delta = event.key === 'ArrowDown' ? 1 : -1;
      const next = (index + delta + faqs.length) % faqs.length;
      buttons[next]?.focus();
    } else if (event.key === 'Home') {
      event.preventDefault();
      buttons[0]?.focus();
    } else if (event.key === 'End') {
      event.preventDefault();
      buttons[faqs.length - 1]?.focus();
    }
  }

  return (
    <LpSection id="lp-faq">
      <LpHeading eyebrow="Вопросы">Частые вопросы</LpHeading>
      <div className="lp-faq-list">
        {faqs.map((item, index) => {
          const isOpen = openIndex === index;
          const buttonId = `lp-faq-button-${index}`;
          const panelId = `lp-faq-panel-${index}`;
          return (
            <div
              key={item.q}
              className={`lp-faq-item${isOpen ? ' is-open' : ''}`}
            >
              <h3>
                <button
                  type="button"
                  id={buttonId}
                  className="lp-faq-question"
                  aria-expanded={isOpen}
                  aria-controls={panelId}
                  onClick={() => toggle(index)}
                  onKeyDown={(event) => onKeyDown(event, index)}
                  ref={(node) => {
                    buttonsRef.current[index] = node;
                  }}
                >
                  <span>{item.q}</span>
                  <span className="lp-faq-icon">
                    <IconChevron />
                  </span>
                </button>
              </h3>
              <div
                id={panelId}
                role="region"
                aria-labelledby={buttonId}
                hidden={!isOpen}
                className="lp-faq-answer"
              >
                <p>{item.a}</p>
              </div>
            </div>
          );
        })}
      </div>
    </LpSection>
  );
}
