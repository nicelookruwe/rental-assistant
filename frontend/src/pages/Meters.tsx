import { useCallback, useEffect, useState } from 'react';
import type { FormEvent } from 'react';

import { api } from '../api';
import { Button, Card, Empty, ErrorNote, Field, Input, Modal, PageHeader, Select, Spinner, Stat, label } from '../components/ui';
import { date, num } from '../format';
import type { Meter, MeterType, Property } from '../types';

interface FormState {
  property_id: string;
  meter_type: MeterType;
  value: string;
  recorded_at: string;
  note: string;
}

function today(): string {
  return new Date().toISOString().slice(0, 10);
}

const emptyForm = (propertyID = ''): FormState => ({
  property_id: propertyID,
  meter_type: 'electricity',
  value: '',
  recorded_at: today(),
  note: '',
});

export default function Meters() {
  const [items, setItems] = useState<Meter[]>([]);
  const [latest, setLatest] = useState<Meter[]>([]);
  const [properties, setProperties] = useState<Property[]>([]);
  const [propertyID, setPropertyID] = useState('');
  const [meterType, setMeterType] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);

  const [open, setOpen] = useState(false);
  const [form, setForm] = useState<FormState>(emptyForm());
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      const data = await api.meters.list({
        property_id: propertyID || undefined,
        meter_type: meterType || undefined,
      });
      setItems(data.items);
      setError(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, [propertyID, meterType]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    void (async () => {
      try {
        const [p, l] = await Promise.all([api.properties.list(), api.meters.latest()]);
        setProperties(p.items);
        setLatest(l.items);
      } catch (err) {
        setError(err);
      }
    })();
  }, []);

  function startCreate() {
    setForm(emptyForm(propertyID || properties[0]?.id || ''));
    setOpen(true);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    const value = Number(form.value.replace(',', '.'));
    if (!form.property_id) {
      setError(new Error('Выберите квартиру'));
      return;
    }
    if (!Number.isFinite(value) || value < 0) {
      setError(new Error('Показание должно быть неотрицательным числом'));
      return;
    }
    setSaving(true);
    try {
      await api.meters.create({
        property_id: form.property_id,
        meter_type: form.meter_type,
        value,
        recorded_at: form.recorded_at,
        note: form.note.trim(),
      });
      setOpen(false);
      setError(null);
      await load();
      const l = await api.meters.latest();
      setLatest(l.items);
    } catch (err) {
      setError(err);
    } finally {
      setSaving(false);
    }
  }

  return (
    <>
      <PageHeader
        title="Счётчики"
        subtitle="История показаний по каждой квартире"
        actions={
          <Button onClick={startCreate} disabled={properties.length === 0}>
            Добавить показание
          </Button>
        }
      />

      <ErrorNote error={error} />

      {latest.length > 0 ? (
        <div className="stats">
          {latest.slice(0, 4).map((m) => (
            <Stat
              key={m.id}
              label={`${label(m.meter_type)} · ${m.property_address}`}
              value={num(m.value)}
              hint={`на ${date(m.recorded_at)}`}
            />
          ))}
        </div>
      ) : null}

      <Card
        title="Последние показания"
        subtitle="Каждое новое значение — отдельная запись, прежние не перезаписываются"
        actions={
          <div className="filters">
            <Select value={propertyID} onChange={(e) => setPropertyID(e.target.value)} aria-label="Квартира">
              <option value="">Все квартиры</option>
              {properties.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.address}
                </option>
              ))}
            </Select>
            <Select value={meterType} onChange={(e) => setMeterType(e.target.value)} aria-label="Тип счётчика">
              <option value="">Все счётчики</option>
              <option value="electricity">Электричество</option>
              <option value="cold_water">Холодная вода</option>
              <option value="hot_water">Горячая вода</option>
              <option value="gas">Газ</option>
              <option value="heating">Отопление</option>
              <option value="other">Прочее</option>
            </Select>
          </div>
        }
      >
        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <Empty
            title="Показаний нет"
            hint={
              properties.length === 0
                ? 'Сначала добавьте квартиру.'
                : 'Внесите первое показание — дальше история будет копиться.'
            }
            action={properties.length > 0 ? <Button onClick={startCreate}>Добавить показание</Button> : undefined}
          />
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Квартира</th>
                  <th>Счётчик</th>
                  <th className="num">Показание</th>
                  <th>Примечание</th>
                </tr>
              </thead>
              <tbody>
                {items.map((meter) => (
                  <tr key={meter.id}>
                    <td className="mono">{date(meter.recorded_at)}</td>
                    <td>{meter.property_address}</td>
                    <td>{label(meter.meter_type)}</td>
                    <td className="num">{num(meter.value)}</td>
                    <td className="muted">{meter.note || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Modal
        open={open}
        title="Новое показание"
        onClose={() => setOpen(false)}
        footer={
          <>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Отмена
            </Button>
            <Button form="meter-form" type="submit" loading={saving}>
              Сохранить
            </Button>
          </>
        }
      >
        <form id="meter-form" className="stack" onSubmit={submit}>
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
          <div className="form-row">
            <Field label="Счётчик">
              <Select
                value={form.meter_type}
                onChange={(e) => setForm({ ...form, meter_type: e.target.value as MeterType })}
              >
                <option value="electricity">Электричество</option>
                <option value="cold_water">Холодная вода</option>
                <option value="hot_water">Горячая вода</option>
                <option value="gas">Газ</option>
                <option value="heating">Отопление</option>
                <option value="other">Прочее</option>
              </Select>
            </Field>
            <Field label="Показание">
              <Input
                value={form.value}
                onChange={(e) => setForm({ ...form, value: e.target.value })}
                inputMode="decimal"
                placeholder="15234,5"
                required
              />
            </Field>
          </div>
          <Field label="Дата снятия">
            <Input
              type="date"
              value={form.recorded_at}
              onChange={(e) => setForm({ ...form, recorded_at: e.target.value })}
              required
            />
          </Field>
          <Field label="Примечание">
            <Input
              value={form.note}
              onChange={(e) => setForm({ ...form, note: e.target.value })}
              placeholder="Показания на 1 ноября"
            />
          </Field>
        </form>
      </Modal>
    </>
  );
}
