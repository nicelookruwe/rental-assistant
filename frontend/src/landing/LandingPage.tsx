import { Link, Route, Routes } from 'react-router-dom';
import type { MouseEvent } from 'react';
import '@fontsource/inter/cyrillic-400.css';
import '@fontsource/inter/latin-400.css';
import '@fontsource/inter/cyrillic-500.css';
import '@fontsource/inter/latin-500.css';
import '@fontsource/inter/cyrillic-600.css';
import '@fontsource/inter/latin-600.css';
import '@fontsource/manrope/cyrillic-600.css';
import '@fontsource/manrope/latin-600.css';
import '@fontsource/manrope/cyrillic-700.css';
import '@fontsource/manrope/latin-700.css';
import '@fontsource/manrope/cyrillic-800.css';
import '@fontsource/manrope/latin-800.css';
import Styleguide from './Styleguide';
import { LpHeading, LpSection } from './ui';
import '../styles/landing/landing.css';
import { HeroSection } from './sections/HeroSection';
import { ProblemSection } from './sections/ProblemSection';
import { FeaturesSection } from './sections/FeaturesSection';
import { HowItWorksSection } from './sections/HowItWorksSection';
import { FaqSection } from './sections/FaqSection';
import { FinalCtaSection } from './sections/FinalCtaSection';

const navLinks = [
  { href: '#lp-features', label: 'Возможности' },
  { href: '#lp-how', label: 'Как работает' },
  { href: '#lp-faq', label: 'Вопросы' },
];

function scrollToHash(event: MouseEvent<HTMLAnchorElement>, href: string) {
  const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (reduce) {
    return;
  }
  event.preventDefault();
  document.querySelector(href)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

function LandingHome() {
  return (
    <div className="lp-root">
      <LpSection className="lp-section-tight">
        <header className="lp-header">
          <Link to="/landing" className="lp-brand">
            Rental Assistant
          </Link>
          <nav className="lp-nav" aria-label="Основная навигация">
            <div className="lp-nav-links">
              {navLinks.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  onClick={(event) => scrollToHash(event, link.href)}
                >
                  {link.label}
                </a>
              ))}
            </div>
            <div className="lp-nav-actions">
              <Link to="/login" className="lp-button lp-button-secondary">
                Войти
              </Link>
              <Link to="/login" className="lp-button lp-button-primary">
                Начать
              </Link>
            </div>
          </nav>
        </header>
      </LpSection>

      <HeroSection />
      <ProblemSection />
      <FeaturesSection />
      <HowItWorksSection />
      <FaqSection />
      <FinalCtaSection />

      <LpSection id="lp-demo" className="lp-demo-placeholder">
        <LpHeading eyebrow="Демо">
          Пространство для следующего этапа
        </LpHeading>
        <p>Здесь появится подробный интерактивный walkthrough продукта.</p>
      </LpSection>

      <LpSection className="lp-footer">
        Rental Assistant · Управление арендой без лишнего шума
      </LpSection>
    </div>
  );
}

export default function LandingPage() {
  return (
    <Routes>
      <Route index element={<LandingHome />} />
      <Route path="styleguide" element={<Styleguide />} />
    </Routes>
  );
}
