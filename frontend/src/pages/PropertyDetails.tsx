import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';

import { api } from '../api';
import { Button, Card, Empty, ErrorNote, PageHeader, Spinner, StatusBadge, label } from '../components/ui';
import { date, fileSize, money, num } from '../format';
import type { Contract, DocumentItem, Meter, Payment, Property } from '../types';

export default function PropertyDetails() {
  const { id = '' } = useParams();
  const navigate = useNavigate();
  const [property, setProperty] = useState<Property | null>(null);
  const [contracts, setContracts] = useState<Contract[]>([]);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [meters, setMeters] = useState<Meter[]>([]);
  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);

  const load = useCallback(async () => {
    try {
      const [p, c, pay, m, d] = await Promise.all([
        api.properties.get(id),
        api.contracts.list({ property_id: id }),
        api.payments.list({ property_id: id }),
        api.meters.list({ property_id: id }),
        api.documents.list({ property_id: id }),
      ]);
      setProperty(p);
      setContracts(c.items);
      setPayments(pay.items);
      setMeters(m.items);
      setDocuments(d.items);
      setError(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  if (loading) return <Spinner />;

  const active = contracts.filter((c) => c.status === 'active');
  const rentIncome = payments
    .filter((p) => p.status === 'paid' && p.type === 'rent')
    .reduce((sum, p) => sum + Number(p.amount), 0);

  async function remove() {
    if (!property) return;
    if (!window.confirm(`Удалить квартиру «${property.address}»?`)) return;
    try {
      await api.properties.remove(property.id);
      navigate('/properties');
    } catch (err) {
      setError(err);
    }
  }

  return (
    <>
      <PageHeader
        title={property?.address ?? 'Квартира'}
        subtitle={
          <Link to="/properties" className="muted">
            ← Все квартиры
          </Link>
        }
        actions={
          <>
            <Button variant="ghost" onClick={() => navigate('/contracts')}>
              Договоры
            </Button>
            <Button variant="danger" onClick={remove}>
              Удалить
            </Button>
          </>
        }
      />

      <ErrorNote error={error} />

      {property ? (
        <Card>
          <dl className="details-grid">
            <div className="detail">
              <dt>Площадь</dt>
              <dd>{num(property.area)} м²</dd>
            </div>
            <div className="detail">
              <dt>Действующих договоров</dt>
              <dd>{active.length}</dd>
            </div>
            <div className="detail">
              <dt>Оплачено аренды</dt>
              <dd>{money(rentIncome)}</dd>
            </div>
            <div className="detail">
              <dt>Арендаторы</dt>
              <dd>{active.map((c) => c.tenant_name).join(', ') || '—'}</dd>
            </div>
            <div className="detail">
              <dt>Добавлена</dt>
              <dd>{date(property.created_at)}</dd>
            </div>
            <div className="detail">
              <dt>Описание</dt>
              <dd>{property.description || '—'}</dd>
            </div>
          </dl>
        </Card>
      ) : null}

      <Card
        title="Договоры"
        actions={
          <Link className="btn btn-ghost" to="/contracts">
            К договорам
          </Link>
        }
      >
        {contracts.length === 0 ? (
          <Empty title="Договоров нет" hint="Создайте договор в разделе «Договоры»." />
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Арендатор</th>
                  <th>Период</th>
                  <th className="num">Аренда</th>
                  <th>Статус</th>
                </tr>
              </thead>
              <tbody>
                {contracts.map((c) => (
                  <tr key={c.id}>
                    <td>{c.tenant_name}</td>
                    <td className="mono">
                      {date(c.start_date)} — {date(c.end_date)}
                    </td>
                    <td className="num">{money(c.rent_amount)}</td>
                    <td>
                      <StatusBadge value={c.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <div className="grid-2">
        <Card title="Последние платежи">
          {payments.length === 0 ? (
            <Empty title="Платежей нет" />
          ) : (
            <div className="list">
              {payments.slice(0, 8).map((p) => (
                <div className="list-item" key={p.id}>
                  <div className="list-main">
                    <div className="list-title">{label(p.type)}</div>
                    <div className="list-meta">
                      {date(p.payment_date)} · {label(p.status)}
                    </div>
                  </div>
                  <span className="list-amount">{money(p.amount)}</span>
                </div>
              ))}
            </div>
          )}
        </Card>

        <Card title="Последние показания">
          {meters.length === 0 ? (
            <Empty title="Показаний нет" />
          ) : (
            <div className="list">
              {meters.slice(0, 8).map((m) => (
                <div className="list-item" key={m.id}>
                  <div className="list-main">
                    <div className="list-title">{label(m.meter_type)}</div>
                    <div className="list-meta">{date(m.recorded_at)}</div>
                  </div>
                  <span className="list-amount">{num(m.value)}</span>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>

      <Card title="Документы">
        {documents.length === 0 ? (
          <Empty title="Документов нет" hint="Договоры, акты и фото загружаются в разделе «Документы»." />
        ) : (
          <div className="list">
            {documents.map((doc) => (
              <div className="list-item" key={doc.id}>
                <div className="list-main">
                  <div className="list-title">{doc.file_name}</div>
                  <div className="list-meta">
                    {label(doc.document_type)} · {fileSize(doc.size_bytes)} · {date(doc.created_at)}
                  </div>
                </div>
                <Button variant="ghost" onClick={() => api.documents.open(doc.id)}>
                  Открыть
                </Button>
              </div>
            ))}
          </div>
        )}
      </Card>
    </>
  );
}
