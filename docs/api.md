# REST API Rental Assistant

Базовый URL: `http://localhost:8080/api`.

## Общие правила

**Аутентификация.** Все маршруты, кроме `POST /api/auth/register`, `POST /api/auth/login`
и `GET /api/health`, требуют заголовок:

```http
Authorization: Bearer <JWT>
```

Токен выдаётся при регистрации и входе, живёт `JWT_TTL` (по умолчанию 168 часов). Владелец данных
берётся только из токена.

**Формат данных.** Тело запроса и ответа — JSON (`Content-Type: application/json`), кроме загрузки
документа (`multipart/form-data`). Даты — `YYYY-MM-DD`, временные метки — RFC 3339.

**Ошибки.** Единый вид:

```json
{ "error": "текст для пользователя" }
```

| Код | Когда |
|-----|-------|
| `400` | некорректный JSON или неверный формат параметра |
| `401` | нет токена, токен истёк или подпись неверна, неверный пароль |
| `404` | объект не найден **или принадлежит другому пользователю** |
| `409` | конфликт: email занят, платёж уже зафиксирован, объект используется |
| `422` | не прошла валидация полей (длина, знак суммы, допустимые значения) |
| `500` | внутренняя ошибка (детали не раскрываются) |

**Списки** возвращаются в виде `{ "items": [...], "count": N }`.

---

## 1. Аутентификация

### POST /api/auth/register

Создаёт пользователя и сразу возвращает токен.

```json
{ "email": "owner@example.com", "password": "минимум 8 символов" }
```

`201 Created`

```json
{
  "user": { "id": "uuid", "email": "owner@example.com", "created_at": "2026-11-19T17:00:00Z" },
  "token": "eyJhbGciOi...",
  "expires_at": "2026-11-26T17:00:00Z"
}
```

Ошибки: `422` — email некорректен или пароль короче 8 символов, `409` — email уже зарегистрирован.

### POST /api/auth/login

```json
{ "email": "owner@example.com", "password": "..." }
```

`200 OK` — тело как у регистрации. `401` — неверная пара email/пароль.

### GET /api/auth/me

`200 OK` — `{ "id", "email", "created_at" }` текущего пользователя. `401` без токена.

---

## 2. Квартиры

Модель:

```json
{
  "id": "uuid",
  "address": "Москва, ул. Тверская, 1",
  "area": 54.5,
  "description": "2 комнаты, 5 этаж",
  "created_at": "2026-11-19T17:00:00Z",
  "updated_at": "2026-11-19T17:00:00Z"
}
```

| Метод и путь | Тело | Ответ | Ошибки |
|--------------|------|-------|--------|
| `GET /api/properties` | — | `200` `{items,count}` | `401` |
| `POST /api/properties` | `{address, area, description?}` | `201` объект | `422` адрес пуст, `area ≤ 0`, `400` |
| `GET /api/properties/{id}` | — | `200` объект | `404`, `401` |
| `PUT /api/properties/{id}` | те же поля | `200` объект | `404`, `422`, `401` |
| `DELETE /api/properties/{id}` | — | `204` | `404`, `409` если есть договоры, `401` |

```bash
curl -X POST http://localhost:8080/api/properties \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"address":"Москва, ул. Тверская, 1","area":54.5,"description":"2 комнаты"}'
```

---

## 3. Арендаторы

```json
{ "id": "uuid", "name": "Иванов Иван Иванович", "phone": "+7 900 000-00-00",
  "email": "ivan@example.com", "created_at": "...", "updated_at": "..." }
```

| Метод и путь | Тело | Ответ | Ошибки |
|--------------|------|-------|--------|
| `GET /api/tenants` | — | `200` `{items,count}` | `401` |
| `POST /api/tenants` | `{name, phone?, email?}` | `201` объект | `422` имя пустое / email некорректен |
| `GET /api/tenants/{id}` | — | `200` объект | `404` |
| `PUT /api/tenants/{id}` | те же поля | `200` объект | `404`, `422` |
| `DELETE /api/tenants/{id}` | — | `204` | `404`, `409` если есть договоры |

---

## 4. Договоры

```json
{
  "id": "uuid",
  "property_id": "uuid",
  "tenant_id": "uuid",
  "property_address": "Москва, ул. Тверская, 1",
  "tenant_name": "Иванов Иван Иванович",
  "start_date": "2026-01-01",
  "end_date": "2026-12-31",
  "rent_amount": 45000,
  "deposit": 45000,
  "status": "active",
  "created_at": "...",
  "updated_at": "..."
}
```

`status`: `draft` | `active` | `terminated` | `expired`. По умолчанию `active`.

| Метод и путь | Параметры / тело | Ответ | Ошибки |
|--------------|------------------|-------|--------|
| `GET /api/contracts?status=&property_id=` | фильтры необязательны | `200` `{items,count}` | `422` некорректный `property_id` |
| `POST /api/contracts` | `{property_id, tenant_id, start_date, end_date, rent_amount, deposit?, status?}` | `201` объект | `422` `end_date ≤ start_date`, суммы < 0, неизвестный статус; `404` чужая квартира/арендатор |
| `GET /api/contracts/{id}` | — | `200` объект | `404` |
| `PUT /api/contracts/{id}` | те же поля | `200` объект | `404`, `422` |
| `DELETE /api/contracts/{id}` | — | `204` | `404`, `409` если есть платежи |

---

## 5. Платежи (аренда и коммунальные)

Платёж — запись журнала: изменить или удалить её нельзя, можно только добавить новую.

```json
{
  "id": "uuid",
  "contract_id": "uuid",
  "property_address": "Москва, ул. Тверская, 1",
  "tenant_name": "Иванов Иван Иванович",
  "amount": 45000,
  "payment_date": "2026-11-10",
  "type": "rent",
  "status": "paid",
  "description": "Аренда за ноябрь",
  "parent_id": null,
  "created_at": "..."
}
```

`type`: `rent` | `utility` | `deposit` | `repair` | `other`.
`status` в ответе: `pending` | `paid` | `overdue` | `canceled`; `overdue` **вычисляется**
(платёж со статусом `pending` и датой в прошлом) и в базе не хранится.

| Метод и путь | Параметры / тело | Ответ | Ошибки |
|--------------|------------------|-------|--------|
| `GET /api/payments?contract_id=&property_id=&type=&status=&from=&to=&limit=&offset=` | `limit` по умолчанию 100 | `200` `{items,count}` | `422` битый UUID или дата |
| `POST /api/payments` | `{contract_id, amount, payment_date, type, status?, description?}` | `201` объект | `422` сумма ≤ 0, неизвестный тип/статус; `404` чужой договор |
| `GET /api/payments/{id}` | — | `200` объект | `404` |
| `POST /api/payments/{id}/settle` | `{payment_date?, amount?, description?}` — все поля необязательны | `201` новая запись со `status: paid` и `parent_id` исходного платежа | `404`; `409` если платёж уже исполнен, отменён или сам является исполнением |

Смысл `settle`: планируемый платёж остаётся в базе как есть, рядом появляется факт оплаты.
Сумма по умолчанию — сумма плана, дата — сегодня.

```bash
# план платежа
curl -X POST http://localhost:8080/api/payments -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"contract_id":"'$C'","amount":45000,"payment_date":"2026-12-10","type":"rent","description":"Аренда за декабрь"}'

# фиксация оплаты (по умолчанию — сегодня, вся сумма)
curl -X POST http://localhost:8080/api/payments/$P/settle -H "Authorization: Bearer $TOKEN"
```

В списке платежей исходный план, у которого есть исполнение, не показывается — виден только факт
оплаты; двойного дохода в отчётах не возникает.

---

## 6. Счётчики

Показание — запись журнала, не перезаписывается: новое значение создаёт новую строку.

```json
{
  "id": "uuid",
  "property_id": "uuid",
  "property_address": "Москва, ул. Тверская, 1",
  "meter_type": "electricity",
  "value": 15234.5,
  "recorded_at": "2026-11-01",
  "note": "",
  "created_at": "..."
}
```

`meter_type`: `cold_water` | `hot_water` | `electricity` | `gas` | `heating` | `other`.

| Метод и путь | Параметры / тело | Ответ | Ошибки |
|--------------|------------------|-------|--------|
| `GET /api/meters?property_id=&meter_type=&limit=` | история, новые сверху | `200` `{items,count}` | `422` битый UUID/тип |
| `GET /api/meters/latest?property_id=` | последнее показание по каждой паре квартира+тип | `200` `{items,count}` | `422` |
| `GET /api/meters/{id}` | — | `200` объект | `404` |
| `POST /api/meters` | `{property_id, meter_type, value, recorded_at, note?}` — `recorded_at` обязателен | `201` объект | `422` значение < 0 или > 1 000 000, неизвестный тип, нет даты, дата в будущем; `404` чужая квартира |

---

## 7. Документы

Файл хранится на сервере (`FILES_DIR`); в базе — метаданные и `file_url`. Скачивание идёт через
API с проверкой владельца.

```json
{
  "id": "uuid",
  "property_id": "uuid",
  "property_address": "Москва, ул. Тверская, 1",
  "file_url": "/api/documents/<id>/file",
  "file_name": "dogovor.pdf",
  "document_type": "contract",
  "size_bytes": 284517,
  "created_at": "..."
}
```

`document_type`: `contract` | `act` | `receipt` | `photo` | `other` (по умолчанию `other`).

| Метод и путь | Параметры / тело | Ответ | Ошибки |
|--------------|------------------|-------|--------|
| `GET /api/documents?property_id=` | фильтр необязателен | `200` `{items,count}` | `422` |
| `POST /api/documents` | `multipart/form-data`: `property_id`, `document_type`, `file` | `201` объект | `422` нет файла / тип не разрешён, `404` чужая квартира |
| `GET /api/documents/{id}/file` | — | `200` файл (`Content-Disposition: attachment`) | `404` чужая квартира или отсутствующий файл |
| `DELETE /api/documents/{id}` | — | `204` (файл удаляется с диска) | `404` |

```bash
curl -X POST http://localhost:8080/api/documents -H "Authorization: Bearer $TOKEN" \
  -F property_id=$PROPERTY -F document_type=contract -F file=@dogovor.pdf
```

Ограничение размера загружаемого файла — 20 МБ.

---

## 8. Напоминания

```json
{ "id": "uuid", "message": "Аренда 45 000,00 ₽ по объекту «…» ожидается 2026-12-10",
  "scheduled_at": "2026-12-10T12:00:00Z", "status": "pending", "created_at": "..." }
```

| Метод и путь | Параметры | Ответ | Ошибки |
|--------------|-----------|-------|--------|
| `GET /api/notifications?status=` | `pending` / `read` | `200` `{items,count}` | `401` |
| `POST /api/notifications/generate` | — | `200` `{created: N, items: [...]}` | `401`, `500` |
| `POST /api/notifications/{id}/read` | — | `200` объект со `status: read` | `404` |

Генерация создаёт напоминания о платежах со сроком в ближайшие 7 дней и договорах, истекающих
в ближайшие 30 дней. Повторный вызов дублей не создаёт (`dedupe_key` = `payment:<id>` /
`contract:<id>`), поэтому `created: 0` — нормальный ответ.

---

## 9. Отчёты

### GET /api/reports/summary

`200 OK`

```json
{
  "properties_count": 2,
  "tenants_count": 2,
  "active_contracts": 2,
  "contracts_ending_soon": 1,
  "overdue_count": 1,
  "month_income": 90000,
  "month_expenses": 3200,
  "month_profit": 86800,
  "overdue_amount": 3200,
  "upcoming_payments": [ /* платежи со сроком в ближайшие 7 дней */ ],
  "overdue_payments": [ /* просроченные */ ],
  "recent_operations": [ /* последние 10 платежей */ ],
  "contracts_ending_soon_list": [ /* договоры, истекающие в 30 дней */ ],
  "monthly": [ { "month": "2026-09", "income": 45000, "expenses": 0 } ]
}
```

`month_income` — сумма фактических оплат (`paid`) за текущий календарный месяц; `month_expenses` —
коммунальные и прочие расходы за тот же период; `monthly` — последние 6 месяцев.

---

## 10. Служебное

### GET /api/health

`200 OK` — `{ "status": "ok" }`. Используется в `HEALTHCHECK` контейнера, авторизация не нужна.

### Статика фронтенда

При заданном `WEB_DIR` с `index.html` сервер отдаёт сборку React на `/` и делает SPA-fallback на
`index.html` для клиентских маршрутов (`/properties/<uuid>` и т. п.). `/api/*` при этом не
перехватывается.
