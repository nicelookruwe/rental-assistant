import type {
  AuthResult,
  Contract,
  ContractInput,
  DocumentItem,
  DocumentType,
  ListResponse,
  Meter,
  MeterInput,
  Notification,
  Payment,
  PaymentInput,
  Property,
  PropertyInput,
  Summary,
  Tenant,
  TenantInput,
  User,
} from './types';

const TOKEN_KEY = 'rental.token';

export const tokenStore = {
  get(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  },
  set(token: string) {
    localStorage.setItem(TOKEN_KEY, token);
  },
  clear() {
    localStorage.removeItem(TOKEN_KEY);
  },
};

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

type Query = Record<string, string | number | undefined | null>;

function buildQuery(params?: Query): string {
  if (!params) return '';
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      search.set(key, String(value));
    }
  });
  const qs = search.toString();
  return qs ? `?${qs}` : '';
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  const token = tokenStore.get();
  if (token) headers.set('Authorization', `Bearer ${token}`);
  if (init.body && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  let response: Response;
  try {
    response = await fetch(`/api${path}`, { ...init, headers });
  } catch {
    throw new ApiError(0, 'Сервер недоступен. Проверьте, запущен ли backend.');
  }

  if (response.status === 401) {
    tokenStore.clear();
    window.dispatchEvent(new Event('auth:expired'));
    throw new ApiError(401, 'Требуется вход в систему');
  }

  if (!response.ok) {
    let message = `Ошибка запроса (${response.status})`;
    try {
      const data = (await response.json()) as { error?: string };
      if (data?.error) message = data.error;
    } catch {
      /* тело не JSON — оставляем общий текст */
    }
    throw new ApiError(response.status, message);
  }

  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

export const api = {
  auth: {
    register: (email: string, password: string) =>
      request<AuthResult>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    login: (email: string, password: string) =>
      request<AuthResult>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    me: () => request<User>('/auth/me'),
  },

  properties: {
    list: () => request<ListResponse<Property>>('/properties'),
    get: (id: string) => request<Property>(`/properties/${id}`),
    create: (input: PropertyInput) =>
      request<Property>('/properties', { method: 'POST', body: JSON.stringify(input) }),
    update: (id: string, input: PropertyInput) =>
      request<Property>(`/properties/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
    remove: (id: string) => request<void>(`/properties/${id}`, { method: 'DELETE' }),
  },

  tenants: {
    list: () => request<ListResponse<Tenant>>('/tenants'),
    create: (input: TenantInput) =>
      request<Tenant>('/tenants', { method: 'POST', body: JSON.stringify(input) }),
    update: (id: string, input: TenantInput) =>
      request<Tenant>(`/tenants/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
    remove: (id: string) => request<void>(`/tenants/${id}`, { method: 'DELETE' }),
  },

  contracts: {
    list: (params?: { status?: string; property_id?: string }) =>
      request<ListResponse<Contract>>(`/contracts${buildQuery(params)}`),
    create: (input: ContractInput) =>
      request<Contract>('/contracts', { method: 'POST', body: JSON.stringify(input) }),
    update: (id: string, input: ContractInput) =>
      request<Contract>(`/contracts/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
    remove: (id: string) => request<void>(`/contracts/${id}`, { method: 'DELETE' }),
  },

  payments: {
    list: (params?: {
      contract_id?: string;
      property_id?: string;
      type?: string;
      status?: string;
      from?: string;
      to?: string;
    }) => request<ListResponse<Payment>>(`/payments${buildQuery(params)}`),
    create: (input: PaymentInput) =>
      request<Payment>('/payments', { method: 'POST', body: JSON.stringify(input) }),
    settle: (id: string, body: { amount?: number; payment_date?: string; description?: string }) =>
      request<Payment>(`/payments/${id}/settle`, { method: 'POST', body: JSON.stringify(body) }),
  },

  meters: {
    list: (params?: { property_id?: string; meter_type?: string; limit?: number }) =>
      request<ListResponse<Meter>>(`/meters${buildQuery(params)}`),
    latest: () => request<ListResponse<Meter>>('/meters/latest'),
    create: (input: MeterInput) =>
      request<Meter>('/meters', { method: 'POST', body: JSON.stringify(input) }),
  },

  documents: {
    list: (params?: { property_id?: string }) =>
      request<ListResponse<DocumentItem>>(`/documents${buildQuery(params)}`),
    upload: (propertyId: string, documentType: DocumentType, file: File) => {
      const form = new FormData();
      form.set('property_id', propertyId);
      form.set('document_type', documentType);
      form.set('file', file);
      return request<DocumentItem>('/documents', { method: 'POST', body: form });
    },
    remove: (id: string) => request<void>(`/documents/${id}`, { method: 'DELETE' }),
    async open(id: string) {
      const token = tokenStore.get();
      const response = await fetch(`/api/documents/${id}/file`, {
        headers: token ? { Authorization: `Bearer ${token}` } : undefined,
      });
      if (!response.ok) throw new ApiError(response.status, 'Не удалось открыть файл');
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      window.open(url, '_blank', 'noopener');
      window.setTimeout(() => URL.revokeObjectURL(url), 60_000);
    },
  },

  notifications: {
    list: (status?: string) => request<ListResponse<Notification>>(`/notifications${buildQuery({ status })}`),
    generate: () =>
      request<ListResponse<Notification>>('/notifications/generate', { method: 'POST' }),
    markRead: (id: string) =>
      request<Notification>(`/notifications/${id}/read`, { method: 'POST' }),
  },

  reports: {
    summary: () => request<Summary>('/reports/summary'),
  },
};
