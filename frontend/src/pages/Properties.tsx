import { useCallback, useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import { Link } from 'react-router-dom';

import { api } from '../api';
import { Button, Card, Empty, ErrorNote, Field, Input, Modal, PageHeader, Spinner, Textarea } from '../components/ui';
import { date, num } from '../format';
import type { Property } from '../types';

interface FormState {
  address: string;
  area: string;
  description: string;
}

const emptyForm: FormState = { address: '', area: '', description: '' };

export default function Properties() {
  const [items, setItems] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Property | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      const data = await api.properties.list();
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

  function startEdit(property: Property) {
    setEditing(property);
    setForm({
      address: property.address,
      area: String(property.area || ''),
      description: property.description ?? '',
    });
    setOpen(true);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    const area = Number(form.area.replace(',', '.'));
    if (!form.address.trim()) {
      setError(new Error('Укажите адрес квартиры'));
      setSaving(false);
      return;
    }
    if (!Number.isFinite(area) || area <= 0) {
      setError(new Error('Площадь должна быть положительным числом'));
      setSaving(false);
      return;
    }
    try {
      const payload = { address: form.address.trim(), area, description: form.description.trim() };
      if (editing) {
        await api.properties.update(editing.id, payload);
      } else {
        await api.properties.create(payload);
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

  async function remove(property: Property) {
    if (!window.confirm(`Удалить квартиру «${property.address}»? Вместе с ней удалятся связанные данные.`)) {
      return;
    }
    try {
      await api.properties.remove(property.id);
      await load();
    } catch (err) {
      setError(err);
    }
  }

  return (
    <>
      <PageHeader
        title="Квартиры"
        subtitle="Объекты, которые вы сдаёте"
        actions={<Button onClick={startCreate}>Добавить квартиру</Button>}
      />

      <ErrorNote error={error} />

      <Card>
        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <Empty
            title="Пока нет ни одной квартиры"
            hint="Добавьте первую квартиру, чтобы завести договор и платежи."
            action={<Button onClick={startCreate}>Добавить квартиру</Button>}
          />
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Адрес</th>
                  <th className="num">Площадь, м²</th>
                  <th>Описание</th>
                  <th>Добавлена</th>
                  <th className="actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                {items.map((property) => (
                  <tr key={property.id}>
                    <td>
                      <Link className="row-link" to={`/properties/${property.id}`}>
                        {property.address}
                      </Link>
                    </td>
                    <td className="num">{num(property.area)}</td>
                    <td className="muted">{property.description || '—'}</td>
                    <td className="muted">{date(property.created_at)}</td>
                    <td className="actions">
                      <Button variant="ghost" onClick={() => startEdit(property)}>
                        Изменить
                      </Button>{' '}
                      <Button variant="danger" onClick={() => remove(property)}>
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
        title={editing ? 'Редактировать квартиру' : 'Новая квартира'}
        onClose={() => setOpen(false)}
        footer={
          <>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Отмена
            </Button>
            <Button form="property-form" type="submit" loading={saving}>
              Сохранить
            </Button>
          </>
        }
      >
        <form id="property-form" className="stack" onSubmit={submit}>
          <Field label="Адрес">
            <Input
              value={form.address}
              onChange={(e) => setForm({ ...form, address: e.target.value })}
              placeholder="Москва, ул. Ленина, 12, кв. 34"
              required
            />
          </Field>
          <Field label="Площадь, м²">
            <Input
              value={form.area}
              onChange={(e) => setForm({ ...form, area: e.target.value })}
              inputMode="decimal"
              placeholder="54,5"
              required
            />
          </Field>
          <Field label="Описание" hint="Этаж, особенности, что важно помнить">
            <Textarea
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              placeholder="3-й этаж, рядом метро, мебель остаётся"
            />
          </Field>
        </form>
      </Modal>
    </>
  );
}
