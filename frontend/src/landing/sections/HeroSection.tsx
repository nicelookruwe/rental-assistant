import { lazy, Suspense } from 'react';
import { Link } from 'react-router-dom';
import { landingConfig } from '../config';
import { ApartmentPanel } from './ApartmentPanel';
import { FloatingCards } from './FloatingCards';
import { BuildingFallback } from '../three/BuildingFallback';
import { WebGLErrorBoundary } from '../three/WebGLErrorBoundary';

const BuildingCanvas = lazy(() => import('../three/BuildingCanvas').then((module) => ({ default: module.BuildingCanvas })));

export function HeroSection() {
  return <section className="lp-hero"><div className="lp-container lp-hero-grid"><div className="lp-hero-copy-block"><h1>{landingConfig.heroTitle}</h1><p className="lp-hero-copy">{landingConfig.heroSubtitle}</p><div className="lp-actions"><Link to="/login" className="lp-button lp-button-primary">{landingConfig.primaryCta}</Link><a href="#lp-demo" className="lp-button lp-button-secondary">{landingConfig.secondaryCta}</a></div></div><div className="lp-experience"><div className="lp-scene"><div className="lp-building-visual"><BuildingFallback /><WebGLErrorBoundary fallback={<BuildingFallback />}><Suspense fallback={<BuildingFallback />}><BuildingCanvas /></Suspense></WebGLErrorBoundary></div><FloatingCards /></div><ApartmentPanel /></div></div></section>;
}
