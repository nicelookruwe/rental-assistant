import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import { api } from '../api';
import { Card, Button, Empty, ErrorNote, PageHeader, Spinner, Stat, StatusBadge, label } from '../components/ui';
import { date, money, monthLabel, num } from '../format';
import type { Notification, Summary } from '../types';

export default function Dashboard() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    try {
      const [s, n] = await Promise.all([api.reports.summary(), api.notifications.list()]);
      setSummary(s);
      setNotifications(n.items);
      setError(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function generateNotifications() {
    setBusy(true);
    try {
      const created = await api.notifications.generate();
      const list = await api.notifications.list();
      setNotifications(list.items);
      setError(null);
      if (created.count === 0) {
        // Напоминаний не появилось — это нормально, просто обновляем список.
      }
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  }

  async function markRead(id: string) {
    try {
      await api.notifications.markRead(id);
      setNotifications((items) => items.map((n) => (n.id === id ? { ...n, status: 'read' } : n)));
    } catch (err) {
      setError(err);
    }
  }

  if (loading) return <Spinner />;

  return (
    <>
      <PageHeader
        title="Обзор"
        subtitle="Сводка по квартирам, договорам и платежам"
        actions={
          <>
            <Link className="btn btn-ghost" to="/payments">
              Платежи
            </Link>
            <Link className="btn btn-primary" to="/properties">
              Квартиры
            </Link>
          </>
        }
      />

      <ErrorNote error={error} />

      {summary ? (
        <>
          <div className="stats">
            <Stat label="Квартиры" value={num(summary.properties_count)} hint={`${num(summary.tenants_count)} арендаторов`} />
            <Stat
              label="Действующие договоры"
              value={num(summary.active_contracts)}
              hint={`Истекают в 30 дней: ${num(summary.contracts_ending_soon)}`}
            />
            <Stat label="Доход за месяц" value={money(summary.month_income)} hint={`Расходы ${money(summary.month_expenses)}`} tone="success" />
            <Stat
              label="Просрочено"
              value={money(summary.overdue_amount)}
              hint={`${num(summary.overdue_count)} платежей`}
              tone={summary.overdue_count > 0 ? 'danger' : 'default'}
            />
          </div>

          <div className="grid-2">
            <Card title="Ближайшие платежи" subtitle="Ожидаемые поступления в ближайшие 14 дней">
              {summary.upcoming_payments.length === 0 ? (
                <Empty title="Ближайших платежей нет" hint="Создайте платёж со статусом «Ожидает» — он появится здесь." />
              ) : (
                <div className="list">
                  {summary.upcoming_payments.map((p) => (
                    <div className="list-item" key={p.id}>
                      <div className="list-main">
                        <div className="list-title">{p.property_address || 'Квартира'}</div>
                        <div className="list-meta">
                          {label(p.type)} · {date(p.payment_date)}
                          {p.tenant_name ? ` · ${p.tenant_name}` : ''}
                        </div>
                      </div>
                      <span className="list-amount">{money(p.amount)}</span>
                    </div>
                  ))}
                </div>
              )}
            </Card>

            <Card title="Просроченные платежи" subtitle="Требуют внимания">
              {summary.overdue_payments.length === 0 ? (
                <Empty title="Просрочек нет" hint="Все платежи в срок." />
              ) : (
                <div className="list">
                  {summary.overdue_payments.map((p) => (
                    <div className="list-item" key={p.id}>
                      <div className="list-main">
                        <div className="list-title">{p.property_address || 'Квартира'}</div>
                        <div className="list-meta">
                          {label(p.type)} · срок {date(p.payment_date)}
                        </div>
                      </div>
                      <span className="list-amount" style={{ color: 'var(--danger)' }}>
                        {money(p.amount)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </div>

          <Card title="Последние операции" actions={<Link className="btn btn-ghost" to="/payments">Все платежи</Link>}>
            {summary.recent_operations.length === 0 ? (
              <Empty title="Операций пока нет" hint="Платежи и оплаты появятся здесь после первого начисления." />
            ) : (
              <div className="table-wrap">
                <table className="table">
                  <thead>
                    <tr>
                      <th>Дата</th>
                      <th>Квартира</th>
                      <th>Тип</th>
                      <th>Описание</th>
                      <th className="num">Сумма</th>
                      <th>Статус</th>
                    </tr>
                  </thead>
                  <tbody>
                    {summary.recent_operations.map((p) => (
                      <tr key={p.id}>
                        <td className="mono">{date(p.payment_date)}</td>
                        <td>{p.property_address || '—'}</td>
                        <td>{label(p.type)}</td>
                        <td className="muted">{p.description || '—'}</td>
                        <td className="num">{money(p.amount)}</td>
                        <td>
                          <StatusBadge value={p.status} />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>

          <div className="grid-2">
            <Card title="Доходы и расходы" subtitle="Последние 6 месяцев">
              {summary.monthly.length === 0 ? (
                <Empty title="Данных для графика пока нет" />
              ) : (
                <>
                  <div className="bars">
                    {summary.monthly.map((point) => {
                      const max = Math.max(
                        ...summary.monthly.flatMap((m) => [m.income, m.expenses]),
                        1,
                      );
                      return (
                        <div className="bar-group" key={point.month}>
                          <div className="bar-pair">
                            <div
                              className="bar"
                              style={{ height: `${Math.max((point.income / max) * 100, 1)}%` }}
                              title={`Доход: ${money(point.income)}`}
                            />
                            <div
                              className="bar expenses"
                              style={{ height: `${Math.max((point.expenses / max) * 100, 1)}%` }}
                              title={`Расходы: ${money(point.expenses)}`}
                            />
                          </div>
                          <span className="bar-label">{monthLabel(point.month)}</span>
                        </div>
                      );
                    })}
                  </div>
                  <div className="legend">
                    <span>Доход</span>
                    <span className="expenses">Расходы</span>
                  </div>
                </>
              )}
            </Card>

            <Card title="Договоры истекают" subtitle="Ближайшие 30 дней">
              {summary.contracts_ending_soon_list.length === 0 ? (
                <Empty title="Скоро ничего не заканчивается" hint="Договоры с близким сроком появятся здесь." />
              ) : (
                <div className="list">
                  {summary.contracts_ending_soon_list.map((c) => (
                    <div className="list-item" key={c.id}>
                      <div className="list-main">
                        <div className="list-title">{c.property_address}</div>
                        <div className="list-meta">
                          {c.tenant_name} · до {date(c.end_date)}
                        </div>
                      </div>
                      <span className="list-amount">{money(c.rent_amount)}</span>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </div>
        </>
      ) : null}

      <Card
        title="Напоминания"
        subtitle="Сроки договоров и ближайшие платежи"
        actions={
          <Button variant="ghost" onClick={generateNotifications} loading={busy}>
            Проверить сроки
          </Button>
        }
      >
        {notifications.length === 0 ? (
          <Empty
            title="Напоминаний нет"
            hint="Нажмите «Проверить сроки» — система соберёт напоминания по договорам и платежам."
          />
        ) : (
          <div className="list">
            {notifications.map((n) => (
              <div className="list-item" key={n.id}>
                <div className="list-main">
                  <div className="list-title" style={{ fontWeight: n.status === 'read' ? 500 : 650 }}>
                    {n.message}
                  </div>
                  <div className="list-meta">напоминание на {date(n.scheduled_at)}</div>
                </div>
                {n.status === 'read' ? (
                  <span className="badge">Прочитано</span>
                ) : (
                  <Button variant="ghost" onClick={() => markRead(n.id)}>
                    Прочитано
                  </Button>
                )}
              </div>
            ))}
          </div>
        )}
      </Card>
    </>
  );
}
