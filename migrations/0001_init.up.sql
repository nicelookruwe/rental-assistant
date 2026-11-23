-- 0001_init.up.sql — начальная схема Rental Assistant.
-- Все сущности привязаны к владельцу (users.id) прямо или через внешние ключи,
-- поэтому изоляция данных проверяется на уровне SQL.

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_not_blank CHECK (length(btrim(email)) > 3)
);
-- Почта уникальна без учёта регистра.
CREATE UNIQUE INDEX users_email_lower_uniq ON users (lower(email));

CREATE TABLE properties (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    address     TEXT          NOT NULL,
    area        NUMERIC(10, 2) NOT NULL,
    description TEXT          NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT properties_address_not_blank CHECK (length(btrim(address)) > 0),
    CONSTRAINT properties_area_positive CHECK (area > 0)
);
CREATE INDEX properties_user_id_idx ON properties (user_id, created_at DESC);

CREATE TABLE tenants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    phone      TEXT        NOT NULL DEFAULT '',
    email      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tenants_name_not_blank CHECK (length(btrim(name)) > 0)
);
CREATE INDEX tenants_user_id_idx ON tenants (user_id, created_at DESC);

CREATE TABLE contracts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID           NOT NULL REFERENCES properties (id) ON DELETE RESTRICT,
    tenant_id   UUID           NOT NULL REFERENCES tenants (id) ON DELETE RESTRICT,
    start_date  DATE           NOT NULL,
    end_date    DATE           NOT NULL,
    rent_amount NUMERIC(14, 2) NOT NULL,
    deposit     NUMERIC(14, 2) NOT NULL DEFAULT 0,
    status      TEXT           NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    CONSTRAINT contracts_period_valid CHECK (end_date > start_date),
    CONSTRAINT contracts_rent_non_negative CHECK (rent_amount >= 0),
    CONSTRAINT contracts_deposit_non_negative CHECK (deposit >= 0),
    CONSTRAINT contracts_status_valid CHECK (status IN ('draft', 'active', 'terminated', 'expired'))
);
CREATE INDEX contracts_property_id_idx ON contracts (property_id, start_date DESC);
CREATE INDEX contracts_tenant_id_idx ON contracts (tenant_id, start_date DESC);
CREATE INDEX contracts_status_idx ON contracts (status, end_date);

-- Платежи — журнал: строки не изменяются и не удаляются приложением.
CREATE TABLE payments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id  UUID           NOT NULL REFERENCES contracts (id) ON DELETE RESTRICT,
    amount       NUMERIC(14, 2) NOT NULL,
    payment_date DATE           NOT NULL,
    type         TEXT           NOT NULL,
    status       TEXT           NOT NULL DEFAULT 'pending',
    description  TEXT           NOT NULL DEFAULT '',
    -- родительский платёж: так планируемый платёж помечается исполненным
    -- без перезаписи исходной записи
    parent_id    UUID           REFERENCES payments (id) ON DELETE RESTRICT,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT now(),
    CONSTRAINT payments_amount_positive CHECK (amount > 0),
    CONSTRAINT payments_type_valid CHECK (type IN ('rent', 'utility', 'deposit', 'repair', 'other')),
    CONSTRAINT payments_status_valid CHECK (status IN ('pending', 'paid', 'overdue', 'canceled'))
);
CREATE INDEX payments_contract_id_idx ON payments (contract_id, payment_date DESC);
CREATE INDEX payments_date_idx ON payments (payment_date DESC);
CREATE INDEX payments_status_idx ON payments (status, payment_date);
CREATE UNIQUE INDEX payments_one_settlement_per_parent ON payments (parent_id)
    WHERE parent_id IS NOT NULL;

-- Показания счётчиков — тоже журнал: только вставка новых значений.
CREATE TABLE meters (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID           NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    property_id UUID           NOT NULL REFERENCES properties (id) ON DELETE CASCADE,
    meter_type  TEXT           NOT NULL,
    value       NUMERIC(14, 3) NOT NULL,
    recorded_at DATE           NOT NULL,
    note        TEXT           NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    CONSTRAINT meters_type_valid CHECK (meter_type IN ('cold_water', 'hot_water', 'electricity', 'gas', 'heating', 'other')),
    CONSTRAINT meters_value_non_negative CHECK (value >= 0)
);
CREATE INDEX meters_property_idx ON meters (property_id, meter_type, recorded_at DESC);
CREATE INDEX meters_user_idx ON meters (user_id, recorded_at DESC);

CREATE TABLE documents (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id   UUID        NOT NULL REFERENCES properties (id) ON DELETE CASCADE,
    file_url      TEXT        NOT NULL,
    document_type TEXT        NOT NULL,
    file_name     TEXT        NOT NULL DEFAULT '',
    size_bytes    BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT documents_type_valid CHECK (document_type IN ('contract', 'act', 'receipt', 'photo', 'other'))
);
CREATE INDEX documents_property_id_idx ON documents (property_id, created_at DESC);

CREATE TABLE notifications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    message      TEXT        NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending',
    dedupe_key   TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT notifications_status_valid CHECK (status IN ('pending', 'sent', 'read', 'canceled')),
    CONSTRAINT notifications_dedupe_uniq UNIQUE (user_id, dedupe_key)
);
CREATE INDEX notifications_user_idx ON notifications (user_id, scheduled_at);

-- updated_at обновляется триггером, а не приложением.
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER properties_set_updated_at BEFORE UPDATE ON properties
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER tenants_set_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER contracts_set_updated_at BEFORE UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Журналы неизменяемы: UPDATE и DELETE запрещены на уровне БД.
CREATE OR REPLACE FUNCTION forbid_mutation() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'таблица % только для добавления: изменение истории запрещено', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER payments_append_only BEFORE UPDATE OR DELETE ON payments
    FOR EACH ROW EXECUTE FUNCTION forbid_mutation();
CREATE TRIGGER meters_append_only BEFORE UPDATE OR DELETE ON meters
    FOR EACH ROW EXECUTE FUNCTION forbid_mutation();
