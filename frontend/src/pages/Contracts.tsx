import { useCallback, useEffect, useMemo, useState } from 'react';
import type { FormEvent } from 'react';

import { api } from '../api';
import { Button, Card, Empty, ErrorNote, Field, Input, Modal, PageHeader, Select, Spinner, StatusBadge } from '../components/ui';
import { date, money } from '../format';
import type { Contract, ContractInput, ContractStatus, Property, Tenant } from '../types';

interface FormState {
  property_id: string;
  tenant_id: string;
  start_date: string;
  end_date: string;
  rent_amount: string;
  deposit: string;
  status: ContractStatus;
}

function today(): string {
  return new Date().toISOString().slice(0, 10);
}

const emptyForm = (): FormState => ({
  property_id: '',
  tenant_id: '',
  start_date: today(),
  end_date: '',
  rent_amount: '',
  deposit: '',
  status: 'active',
});

export default function Contracts() {
  const [items, setItems] = useState<Contract[]>([]);
  const [properties, setProperties] = useState<Property[]>([]);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [status, setStatus] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Contract | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      const data = await api.contracts.list(status ? { status } : undefined);
      setItems(data.items);
      setError(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, [status]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    void (async () => {
      try {
        const [p, t] = await Promise.all([api.properties.list(), api.tenants.list()]);
        setProperties(p.items);
        setTenants(t.items);
      } catch (err) {
        setError(err);
      }
    })();
  }, []);

  const canCreate = properties.length > 0 && tenants.length > 0;

  function startCreate() {
    setEditing(null);
    const base = emptyForm();
    setForm({
      ...base,
      property_id: properties[0]?.id ?? '',
      tenant_id: tenants[0]?.id ?? '',
    });
    setOpen(true);
  }

  function startEdit(contract: Contract) {
    setEditing(contract);
    setForm({
      property_id: contract.property_id,
      tenant_id: contract.tenant_id,
      start_date: contract.start_date,
      end_date: contract.end_date,
      rent_amount: String(contract.rent_amount),
      deposit: String(contract.deposit ?? ''),
      status: contract.status,
    });
    setOpen(true);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    const rent = Number(form.rent_amount.replace(',', '.'));
    const deposit = form.deposit ? Number(form.deposit.replace(',', '.')) : 0;

    if (!form.property_id || !form.tenant_id) {
      setError(new Error('Выберите квартиру и арендатора'));
      return;
    }
    if (!Number.isFinite(rent) || rent <= 0) {
      setError(new Error('Арендная плата должна быть больше нуля'));
      return;
    }
    if (!Number.isFinite(deposit) || deposit < 0) {
      setError(new Error('Депозит не может быть отрицательным'));
      return;
    }
    if (form.end_date && form.end_date <= form.start_date) {
      setError(new Error('Дата окончания должна быть позже даты начала'));
      return;
    }

    const payload: ContractInput = {
      property_id: form.property_id,
      tenant_id: form.tenant_id,
      start_date: form.start_date,
      end_date: form.end_date,
      rent_amount: rent,
      deposit,
      status: form.status,
    };

    setSaving(true);
    try {
      if (editing) {
        await api.contracts.update(editing.id, payload);
      } else {
        await api.contracts.create(payload);
      }
      setOpen(false);
      setError(null);
      await load();
    } catch (err) {
      setError(err);
    } finally {
      setSaving(false);
    }
  }

  async function remove(contract: Contract) {
    if (!window.confirm(`Удалить договор с «${contract.tenant_name}»?`)) return;
    try {
      await api.contracts.remove(contract.id);
      await load();
    } catch (err) {
      setError(err);
    }
  }

  const totalRent = useMemo(
    () => items.filter((c) => c.status === 'active').reduce((sum, c) => sum + Number(c.rent_amount), 0),
    [items],
  );

  return (
    <>
      <PageHeader
        title="Договоры"
        subtitle="Сроки, ставки и депозиты по каждой квартире"
        actions={<Button onClick={startCreate}>Новый договор</Button>}
      />

      <ErrorNote error={error} />

      <Card
        title={`Всего договоров: ${items.length}`}
        subtitle={`Аренда по действующим: ${money(totalRent)} в месяц`}
        actions={
          <Select value={status} onChange={(e) => setStatus(e.target.value)} aria-label="Фильтр по статусу">
            <option value="">Все статусы</option>
            <option value="active">Действующие</option>
            <option value="draft">Черновики</option>
            <option value="expired">Истёкшие</option>
            <option value="terminated">Расторгнутые</option>
          </Select>
        }
      >
        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <Empty
            title="Договоров нет"
            hint={
              canCreate
                ? 'Создайте договор — он свяжет квартиру и арендатора.'
                : 'Сначала добавьте квартиру и арендатора, затем заключите договор.'
            }
            action={canCreate ? <Button onClick={startCreate}>Новый договор</Button> : undefined}
          />
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Квартира</th>
                  <th>Арендатор</th>
                  <th>Период</th>
                  <th className="num">Аренда</th>
                  <th className="num">Депозит</th>
                  <th>Статус</th>
                  <th className="actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                {items.map((contract) => (
                  <tr key={contract.id}>
                    <td>{contract.property_address}</td>
                    <td>{contract.tenant_name}</td>
                    <td className="mono">
                      {date(contract.start_date)} — {date(contract.end_date)}
                    </td>
                    <td className="num">{money(contract.rent_amount)}</td>
                    <td className="num">{money(contract.deposit)}</td>
                    <td>
                      <StatusBadge value={contract.status} />
                    </td>
                    <td className="actions">
                      <Button variant="ghost" onClick={() => startEdit(contract)}>
                        Изменить
                      </Button>{' '}
                      <Button variant="danger" onClick={() => remove(contract)}>
                        Удалить
                      </Button>
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
        title={editing ? 'Редактировать договор' : 'Новый договор'}
        onClose={() => setOpen(false)}
        footer={
          <>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Отмена
            </Button>
            <Button form="contract-form" type="submit" loading={saving}>
              Сохранить
            </Button>
          </>
        }
      >
        <form id="contract-form" className="stack" onSubmit={submit}>
          <Field label="Квартира">
            <Select
              value={form.property_id}
              onChange={(e) => setForm({ ...form, property_id: e.target.value })}
              required
            >
              <option value="">Выберите квартиру</option>
              {properties.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.address}
                </option>
              ))}
            </Select>
          </Field>
          <Field label="Арендатор">
            <Select value={form.tenant_id} onChange={(e) => setForm({ ...form, tenant_id: e.target.value })} required>
              <option value="">Выберите арендатора</option>
              {tenants.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </Select>
          </Field>
          <div className="form-row">
            <Field label="Начало">
              <Input
                type="date"
                value={form.start_date}
                onChange={(e) => setForm({ ...form, start_date: e.target.value })}
                required
              />
            </Field>
            <Field label="Окончание">
              <Input
                type="date"
                value={form.end_date}
                onChange={(e) => setForm({ ...form, end_date: e.target.value })}
                required
              />
            </Field>
          </div>
          <div className="form-row">
            <Field label="Аренда в месяц, ₽">
              <Input
                value={form.rent_amount}
                onChange={(e) => setForm({ ...form, rent_amount: e.target.value })}
                inputMode="decimal"
                placeholder="45 000"
                required
              />
            </Field>
            <Field label="Депозит, ₽">
              <Input
                value={form.deposit}
                onChange={(e) => setForm({ ...form, deposit: e.target.value })}
                inputMode="decimal"
                placeholder="45 000"
              />
            </Field>
          </div>
          <Field label="Статус">
            <Select
              value={form.status}
              onChange={(e) => setForm({ ...form, status: e.target.value as ContractStatus })}
            >
              <option value="active">Действует</option>
              <option value="draft">Черновик</option>
              <option value="expired">Истёк</option>
              <option value="terminated">Расторгнут</option>
            </Select>
          </Field>
        </form>
      </Modal>
    </>
  );
}
