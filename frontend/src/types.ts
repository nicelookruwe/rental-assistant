export interface User {
  id: string;
  email: string;
  created_at: string;
}

export interface AuthResult {
  user: User;
  token: string;
  expires_at: string;
}

export interface Property {
  id: string;
  address: string;
  area: number;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface PropertyInput {
  address: string;
  area: number;
  description: string;
}

export interface Tenant {
  id: string;
  name: string;
  phone: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface TenantInput {
  name: string;
  phone: string;
  email: string;
}

export type ContractStatus = 'draft' | 'active' | 'terminated' | 'expired';

export interface Contract {
  id: string;
  property_id: string;
  tenant_id: string;
  property_address: string;
  tenant_name: string;
  start_date: string;
  end_date: string;
  rent_amount: number;
  deposit: number;
  status: ContractStatus;
  created_at: string;
  updated_at: string;
}

export interface ContractInput {
  property_id: string;
  tenant_id: string;
  start_date: string;
  end_date: string;
  rent_amount: number;
  deposit: number;
  status: ContractStatus;
}

export type PaymentType = 'rent' | 'utility' | 'deposit' | 'repair' | 'other';
export type PaymentStatus = 'pending' | 'paid' | 'overdue' | 'canceled';

export interface Payment {
  id: string;
  contract_id: string;
  property_address: string;
  tenant_name: string;
  amount: number;
  payment_date: string;
  type: PaymentType;
  status: PaymentStatus;
  description: string;
  parent_id?: string;
  created_at: string;
}

export interface PaymentInput {
  contract_id: string;
  amount: number;
  payment_date: string;
  type: PaymentType;
  status: PaymentStatus;
  description: string;
}

export type MeterType = 'electricity' | 'cold_water' | 'hot_water' | 'gas' | 'heating' | 'other';

export interface Meter {
  id: string;
  property_id: string;
  property_address: string;
  meter_type: MeterType;
  value: number;
  recorded_at: string;
  note: string;
  created_at: string;
}

export interface MeterInput {
  property_id: string;
  meter_type: MeterType;
  value: number;
  recorded_at: string;
  note: string;
}

export type DocumentType = 'contract' | 'act' | 'receipt' | 'photo' | 'other';

export interface DocumentItem {
  id: string;
  property_id: string;
  property_address: string;
  file_url: string;
  file_name: string;
  document_type: DocumentType;
  size_bytes: number;
  created_at: string;
}

export interface Notification {
  id: string;
  message: string;
  scheduled_at: string;
  status: 'pending' | 'read';
  created_at: string;
}

export interface MonthPoint {
  month: string;
  income: number;
  expenses: number;
}

export interface Summary {
  properties_count: number;
  tenants_count: number;
  active_contracts: number;
  contracts_ending_soon: number;
  overdue_count: number;
  month_income: number;
  month_expenses: number;
  month_profit: number;
  overdue_amount: number;
  upcoming_payments: Payment[];
  overdue_payments: Payment[];
  recent_operations: Payment[];
  contracts_ending_soon_list: Contract[];
  monthly: MonthPoint[];
}

export interface ListResponse<T> {
  items: T[];
  count: number;
}
