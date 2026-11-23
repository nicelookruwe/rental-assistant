import { useCallback, useEffect, useMemo, useState } from 'react';
import type { FormEvent } from 'react';

import { api } from '../api';
import {
  Button,
  Card,
  Empty,
  ErrorNote,
  Field,
  Input,
  Modal,
  PageHeader,
  Select,
  Spinner,
  Stat,
  StatusBadge,
  label,
} from '../components/ui';
import { date, money } from '../format';
import type { Contract, Payment, PaymentType } from '../types';

interface FormState {
  contract_id: string;
  amount: string;
  payment_date: string;
  type: PaymentType;
  status: 'pending' | 'paid' | 'canceled';
  description: string;
}

function today(): string {
  return new Date().toISOString().slice(0, 10);
}

const emptyForm = (contractID = ''): FormState => ({
  contract_id: contractID,
  amount: '',
  payment_date: today(),
  type: 'rent',
  status: 'paid',
  description: '',
});

export default function Payments() {
  const [items, setItems] = useState<Payment[]>([]);
  const [contracts, setContracts] = useState<Contract[]>([]);
  const [type, setType] = useState('');
  const [status, setStatus] = useState('');
  const [contractID, setContractID] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);

  const [open, setOpen] = useState(false);
  const [form, setForm] = useState<FormState>(emptyForm());
  const [saving, setSaving] = useState(false);

  const [settling, setSettling] = useState<Payment | null>(null);
  const [settleAmount, setSettleAmount] = useState('');
  const [settleDate, setSettleDate] = useState(today());

  const load = useCallback(async () => {
    try {
      const data = await api.payments.list({
        type: type || undefined,
        status: status || undefined,
        contract_id: contractID || undefined,
      });
      setItems(data.items);
      setError(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, [type, status, contractID]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    void (async () => {
      try {
        const data = await api.contracts.list();
        setContracts(data.items);
      } catch (err) {
        setError(err);
      }
    })();
  }, []);

  const totals = useMemo(() => {
    return items.reduce(
      (acc, p) => {
        if (p.status === 'paid') acc.paid += Number(p.amount);
        if (p.status === 'pending') acc.pending += Number(p.amount);
        if (p.status === 'overdue') acc.overdue += Number(p.amount);
        return acc;
      },
      { paid: 0, pending: 0, overdue: 0 },
    );
  }, [items]);

  function startCreate() {
    setForm(emptyForm(contracts.find((c) => c.status === 'active')?.id ?? contracts[0]?.id ?? ''));
    setOpen(true);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    const amount = Number(form.amount.replace(',', '.'));
    if (!form.contract_id) {
      setError(new Error('Выберите договор'));
      return;
    }
    if (!Number.isFinite(amount) || amount <= 0) {
      setError(new Error('Сумма должна быть больше нуля'));
      return;
    }
    setSaving(true);
    try {
      await api.payments.create({
        contract_id: form.contract_id,
        amount,
        payment_date: form.payment_date,
        type: form.type,
        status: form.status,
        description: form.description.trim(),
      });
      setOpen(false);
      setError(null);
      await load();
    } catch (err) {
      setError(err);
    } finally {
      setSaving(false);
    }
  }

  function startSettle(payment: Payment) {
    setSettling(payment);
    setSettleAmount(String(payment.amount));
    setSettleDate(today());
  }

  async function confirmSettle(event: FormEvent) {
    event.preventDefault();
    if (!settling) return;
    const amount = Number(settleAmount.replace(',', '.'));
    setSaving(true);
    try {
      await api.payments.settle(settling.id, {
        amount: Number.isFinite(amount) && amount > 0 ? amount : undefined,
        payment_date: settleDate,
      });
      setSettling(null);
      setError(null);
      await load();
    } catch (err) {
      setError(err);
    } finally {
      setSaving(false);
    }
  }

  return (
    <>
      <PageHeader
        title="Платежи"
        subtitle="Аренда, коммунальные и депозиты — журнал операций"
        actions={
          <Button onClick={startCreate} disabled={contracts.length === 0}>
            Добавить платёж
          </Button>
        }
      />

      <ErrorNote error={error} />

      <div className="stats">
        <Stat label="Оплачено" value={money(totals.paid)} tone="success" />
        <Stat label="Ожидается" value={money(totals.pending)} />
        <Stat label="Просрочено" value={money(totals.overdue)} tone={totals.overdue > 0 ? 'danger' : 'default'} />
        <Stat label="Записей" value={items.length} hint="История не перезаписывается" />
      </div>

      <Card
        title="Журнал платежей"
        actions={
          <div className="filters">
            <Select value={contractID} onChange={(e) => setContractID(e.target.value)} aria-label="Договор">
              <option value="">Все договоры</option>
              {contracts.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.property_address} · {c.tenant_name}
                </option>
              ))}
            </Select>
            <Select value={type} onChange={(e) => setType(e.target.value)} aria-label="Тип">
              <option value="">Все типы</option>
              <option value="rent">Аренда</option>
              <option value="utility">Коммунальные</option>
              <option value="deposit">Депозит</option>
              <option value="repair">Ремонт</option>
              <option value="other">Прочее</option>
            </Select>
            <Select value={status} onChange={(e) => setStatus(e.target.value)} aria-label="Статус">
              <option value="">Все статусы</option>
              <option value="paid">Оплачен</option>
              <option value="pending">Ожидает</option>
              <option value="overdue">Просрочен</option>
              <option value="canceled">Отменён</option>
            </Select>
          </div>
        }
      >
        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <Empty
            title="Платежей нет"
            hint={
              contracts.length === 0
                ? 'Сначала создайте договор, затем добавляйте начисления.'
                : 'Добавьте первый платёж: аренду, коммунальные или депозит.'
            }
            action={contracts.length > 0 ? <Button onClick={startCreate}>Добавить платёж</Button> : undefined}
          />
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Квартира</th>
                  <th>Арендатор</th>
                  <th>Тип</th>
                  <th>Описание</th>
                  <th className="num">Сумма</th>
                  <th>Статус</th>
                  <th className="actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                {items.map((payment) => (
                  <tr key={payment.id}>
                    <td className="mono">{date(payment.payment_date)}</td>
                    <td>{payment.property_address || '—'}</td>
                    <td>{payment.tenant_name || '—'}</td>
                    <td>{label(payment.type)}</td>
                    <td className="muted">{payment.description || '—'}</td>
                    <td className="num">{money(payment.amount)}</td>
                    <td>
                      <StatusBadge value={payment.status} />
                    </td>
                    <td className="actions">
                      {payment.status === 'pending' || payment.status === 'overdue' ? (
                        <Button variant="ghost" onClick={() => startSettle(payment)}>
                          Отметить оплату
                        </Button>
                      ) : payment.parent_id ? (
                        <span className="muted mono" title="Запись создана при фиксации оплаты">
                          оплата
                        </span>
                      ) : (
                        <span className="muted">—</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Modal
        open={open}
        title="Новый платёж"
        onClose={() => setOpen(false)}
        footer={
          <>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Отмена
            </Button>
            <Button form="payment-form" type="submit" loading={saving}>
              Сохранить
            </Button>
          </>
        }
      >
        <form id="payment-form" className="stack" onSubmit={submit}>
          <Field label="Договор">
            <Select
              value={form.contract_id}
              onChange={(e) => setForm({ ...form, contract_id: e.target.value })}
              required
            >
              <option value="">Выберите договор</option>
              {contracts.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.property_address} · {c.tenant_name}
                </option>
              ))}
            </Select>
          </Field>
          <div className="form-row">
            <Field label="Сумма, ₽">
              <Input
                value={form.amount}
                onChange={(e) => setForm({ ...form, amount: e.target.value })}
                inputMode="decimal"
                placeholder="45 000"
                required
              />
            </Field>
            <Field label="Дата">
              <Input
                type="date"
                value={form.payment_date}
                onChange={(e) => setForm({ ...form, payment_date: e.target.value })}
                required
              />
            </Field>
          </div>
          <div className="form-row">
            <Field label="Тип">
              <Select
                value={form.type}
                onChange={(e) => setForm({ ...form, type: e.target.value as PaymentType })}
              >
                <option value="rent">Аренда</option>
                <option value="utility">Коммунальные</option>
                <option value="deposit">Депозит</option>
                <option value="repair">Ремонт</option>
                <option value="other">Прочее</option>
              </Select>
            </Field>
            <Field label="Статус" hint="«Просрочен» система ставит сама по дате">
              <Select
                value={form.status}
                onChange={(e) => setForm({ ...form, status: e.target.value as FormState['status'] })}
              >
                <option value="paid">Оплачен</option>
                <option value="pending">Ожидает</option>
                <option value="canceled">Отменён</option>
              </Select>
            </Field>
          </div>
          <Field label="Описание">
            <Input
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              placeholder="Аренда за ноябрь"
            />
          </Field>
        </form>
      </Modal>

      <Modal
        open={settling !== null}
        title="Фиксация оплаты"
        onClose={() => setSettling(null)}
        footer={
          <>
            <Button variant="ghost" onClick={() => setSettling(null)}>
              Отмена
            </Button>
            <Button form="settle-form" type="submit" loading={saving}>
              Подтвердить оплату
            </Button>
          </>
        }
      >
        <form id="settle-form" className="stack" onSubmit={confirmSettle}>
          <p className="muted" style={{ margin: 0 }}>
            {settling ? `${settling.property_address} · ${label(settling.type)} · срок ${date(settling.payment_date)}` : ''}
          </p>
          <div className="form-row">
            <Field label="Дата оплаты">
              <Input
                type="date"
                value={settleDate}
                onChange={(e) => setSettleDate(e.target.value)}
                required
              />
            </Field>
            <Field label="Сумма, ₽">
              <Input
                value={settleAmount}
                onChange={(e) => setSettleAmount(e.target.value)}
                inputMode="decimal"
                required
              />
            </Field>
          </div>
          <p className="field-hint" style={{ margin: 0 }}>
            Оплата сохраняется отдельной записью: исходное начисление остаётся в истории.
          </p>
        </form>
      </Modal>
    </>
  );
}
