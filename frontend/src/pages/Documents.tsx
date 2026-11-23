import { useCallback, useEffect, useState } from 'react';
import type { ChangeEvent, FormEvent } from 'react';

import { api } from '../api';
import { Button, Card, Empty, ErrorNote, Field, Modal, PageHeader, Select, Spinner, StatusBadge } from '../components/ui';
import { date, fileSize } from '../format';
import type { DocumentItem, DocumentType, Property } from '../types';

export default function Documents() {
  const [items, setItems] = useState<DocumentItem[]>([]);
  const [properties, setProperties] = useState<Property[]>([]);
  const [propertyID, setPropertyID] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);

  const [open, setOpen] = useState(false);
  const [uploadProperty, setUploadProperty] = useState('');
  const [uploadType, setUploadType] = useState<DocumentType>('contract');
  const [file, setFile] = useState<File | null>(null);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      const data = await api.documents.list(propertyID ? { property_id: propertyID } : undefined);
      setItems(data.items);
      setError(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, [propertyID]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    void (async () => {
      try {
        const data = await api.properties.list();
        setProperties(data.items);
      } catch (err) {
        setError(err);
      }
    })();
  }, []);

  function startUpload() {
    setUploadProperty(propertyID || properties[0]?.id || '');
    setUploadType('contract');
    setFile(null);
    setOpen(true);
  }

  function pickFile(event: ChangeEvent<HTMLInputElement>) {
    setFile(event.target.files?.[0] ?? null);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!uploadProperty) {
      setError(new Error('Выберите квартиру'));
      return;
    }
    if (!file) {
      setError(new Error('Выберите файл'));
      return;
    }
    setSaving(true);
    try {
      await api.documents.upload(uploadProperty, uploadType, file);
      setOpen(false);
      setError(null);
      await load();
    } catch (err) {
      setError(err);
    } finally {
      setSaving(false);
    }
  }

  async function remove(doc: DocumentItem) {
    if (!window.confirm(`Удалить файл «${doc.file_name}»?`)) return;
    try {
      await api.documents.remove(doc.id);
      await load();
    } catch (err) {
      setError(err);
    }
  }

  async function openFile(doc: DocumentItem) {
    try {
      await api.documents.open(doc.id);
    } catch (err) {
      setError(err);
    }
  }

  return (
    <>
      <PageHeader
        title="Документы"
        subtitle="Договоры, акты, чеки и фотографии по квартирам"
        actions={
          <Button onClick={startUpload} disabled={properties.length === 0}>
            Загрузить файл
          </Button>
        }
      />

      <ErrorNote error={error} />

      <Card
        title={`Файлов: ${items.length}`}
        subtitle="Доступны только вам: файлы отдаются по вашему токену"
        actions={
          <Select value={propertyID} onChange={(e) => setPropertyID(e.target.value)} aria-label="Квартира">
            <option value="">Все квартиры</option>
            {properties.map((p) => (
              <option key={p.id} value={p.id}>
                {p.address}
              </option>
            ))}
          </Select>
        }
      >
        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <Empty
            title="Документов нет"
            hint={
              properties.length === 0
                ? 'Сначала добавьте квартиру.'
                : 'Загрузите договор, акт приёма-передачи или чек — файлы будут привязаны к квартире.'
            }
            action={properties.length > 0 ? <Button onClick={startUpload}>Загрузить файл</Button> : undefined}
          />
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Квартира</th>
                  <th>Тип</th>
                  <th>Файл</th>
                  <th className="num">Размер</th>
                  <th className="actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                {items.map((doc) => (
                  <tr key={doc.id}>
                    <td className="mono">{date(doc.created_at)}</td>
                    <td>{doc.property_address}</td>
                    <td>
                      <StatusBadge value={doc.document_type} />
                    </td>
                    <td className="mono">{doc.file_name}</td>
                    <td className="num">{fileSize(doc.size_bytes)}</td>
                    <td className="actions">
                      <Button variant="ghost" onClick={() => openFile(doc)}>
                        Открыть
                      </Button>{' '}
                      <Button variant="danger" onClick={() => remove(doc)}>
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
        title="Загрузка документа"
        onClose={() => setOpen(false)}
        footer={
          <>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Отмена
            </Button>
            <Button form="document-form" type="submit" loading={saving}>
              Загрузить
            </Button>
          </>
        }
      >
        <form id="document-form" className="stack" onSubmit={submit}>
          <Field label="Квартира">
            <Select value={uploadProperty} onChange={(e) => setUploadProperty(e.target.value)} required>
              <option value="">Выберите квартиру</option>
              {properties.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.address}
                </option>
              ))}
            </Select>
          </Field>
          <Field label="Тип документа">
            <Select value={uploadType} onChange={(e) => setUploadType(e.target.value as DocumentType)}>
              <option value="contract">Договор</option>
              <option value="act">Акт</option>
              <option value="receipt">Чек</option>
              <option value="photo">Фото</option>
              <option value="other">Прочее</option>
            </Select>
          </Field>
          <Field label="Файл" hint="PDF, изображения или документ Office — до 20 МБ">
            <input className="input" type="file" onChange={pickFile} required />
          </Field>
        </form>
      </Modal>
    </>
  );
}
