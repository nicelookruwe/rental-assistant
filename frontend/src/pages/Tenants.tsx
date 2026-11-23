import { useCallback, useEffect, useState } from 'react';
import type { FormEvent } from 'react';

import { api } from '../api';
import { Button, Card, Empty, ErrorNote, Field, Input, Modal, PageHeader, Spinner } from '../components/ui';
import type { Tenant } from '../types';

interface FormState {
  name: string;
  phone: string;
  email: string;
}

const emptyForm: FormState = { name: '', phone: '', email: '' };

export default function Tenants() {
  const [items, setItems] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Tenant | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      const data = await api.tenants.list();
      setItems(data.items);
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

  function startCreate() {
    setEditing(null);
    setForm(emptyForm);
    setOpen(true);
  }

  function startEdit(tenant: Tenant) {
    setEditing(tenant);
    setForm({ name: tenant.name, phone: tenant.phone ?? '', email: tenant.email ?? '' });
    setOpen(true);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!form.name.trim()) {
      setError(new Error('Укажите имя арендатора'));
      return;
    }
    setSaving(true);
    try {
      const payload = {
        name: form.name.trim(),
        phone: form.phone.trim(),
        email: form.email.trim(),
      };
      if (editing) {
        await api.tenants.update(editing.id, payload);
      } else {
        await api.tenants.create(payload);
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

  async function remove(tenant: Tenant) {
    if (!window.confirm(`Удалить арендатора «${tenant.name}»?`)) return;
    try {
      await api.tenants.remove(tenant.id);
      await load();
    } catch (err) {
      setError(err);
    }
  }

  return (
    <>
      <PageHeader
        title="Арендаторы"
        subtitle="Контакты людей, которые снимают ваши квартиры"
        actions={<Button onClick={startCreate}>Добавить арендатора</Button>}
      />

      <ErrorNote error={error} />

      <Card>
        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <Empty
            title="Арендаторов пока нет"
            hint="Добавьте арендатора, чтобы заключить с ним договор."
            action={<Button onClick={startCreate}>Добавить арендатора</Button>}
          />
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Имя</th>
                  <th>Телефон</th>
                  <th>Email</th>
                  <th className="actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                {items.map((tenant) => (
                  <tr key={tenant.id}>
                    <td>
                      <span className="row-link">{tenant.name}</span>
                    </td>
                    <td className="mono">{tenant.phone || '—'}</td>
                    <td className="muted">{tenant.email || '—'}</td>
                    <td className="actions">
                      <Button variant="ghost" onClick={() => startEdit(tenant)}>
                        Изменить
                      </Button>{' '}
                      <Button variant="danger" onClick={() => remove(tenant)}>
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
        title={editing ? 'Редактировать арендатора' : 'Новый арендатор'}
        onClose={() => setOpen(false)}
        footer={
          <>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Отмена
            </Button>
            <Button form="tenant-form" type="submit" loading={saving}>
              Сохранить
            </Button>
          </>
        }
      >
        <form id="tenant-form" className="stack" onSubmit={submit}>
          <Field label="Имя">
            <Input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Иванов Иван Иванович"
              required
            />
          </Field>
          <Field label="Телефон">
            <Input
              value={form.phone}
              onChange={(e) => setForm({ ...form, phone: e.target.value })}
              placeholder="+7 900 000-00-00"
              inputMode="tel"
            />
          </Field>
          <Field label="Email">
            <Input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              placeholder="tenant@example.com"
            />
          </Field>
        </form>
      </Modal>
    </>
  );
}
