import { Link, Route, Routes } from 'react-router-dom';
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

function LandingHome() {
  return <div className="lp-root"><LpSection className="lp-section-tight"><header className="lp-header"><Link to="/landing" className="lp-brand">Rental Assistant</Link><nav className="lp-nav" aria-label="Основная навигация"><Link to="/login" className="lp-button lp-button-secondary">Войти</Link><Link to="/login" className="lp-button lp-button-primary">Начать</Link></nav></header></LpSection><HeroSection /><LpSection id="lp-demo" className="lp-demo-placeholder"><LpHeading eyebrow="Демо">Пространство для следующего этапа</LpHeading><p>Здесь появится подробный интерактивный walkthrough продукта.</p></LpSection><LpSection className="lp-footer"><div className="lp-container">Rental Assistant · Управление арендой без лишнего шума</div></LpSection></div>;
}

export default function LandingPage() { return <Routes><Route index element={<LandingHome />} /><Route path="styleguide" element={<Styleguide />} /></Routes>; }
