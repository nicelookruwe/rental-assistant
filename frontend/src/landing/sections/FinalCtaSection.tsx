import { Link } from 'react-router-dom';
import { LpCard, LpHeading, LpSection } from '../ui';

export function FinalCtaSection() {
  return (
    <LpSection className="lp-final-cta-section">
      <LpCard className="lp-final-cta">
        <LpHeading eyebrow="Готовы начать">
          Управляйте арендой, а не беспорядком
        </LpHeading>
        <div className="lp-actions lp-final-actions">
          <Link to="/login" className="lp-button lp-button-primary">
            Начать
          </Link>
        </div>
      </LpCard>
    </LpSection>
  );
}
