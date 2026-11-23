// Package reports — сводная аналитика для дашборда.
package reports

import (
	"github.com/ailina/rental-assistant/internal/contracts"
	"github.com/ailina/rental-assistant/internal/payments"
)

// MonthPoint — доход и расходы за календарный месяц.
type MonthPoint struct {
	Month    string  `json:"month"` // ГГГГ-ММ
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
}

// Counters — количественные показатели портфеля.
type Counters struct {
	PropertiesCount     int `json:"properties_count"`
	TenantsCount        int `json:"tenants_count"`
	ActiveContracts     int `json:"active_contracts"`
	ContractsEndingSoon int `json:"contracts_ending_soon"`
	OverdueCount        int `json:"overdue_count"`
}

// Money — суммы за период.
type Money struct {
	MonthIncome   float64 `json:"month_income"`
	MonthExpenses float64 `json:"month_expenses"`
	MonthProfit   float64 `json:"month_profit"`
	OverdueAmount float64 `json:"overdue_amount"`
}

// Summary — то, что показывает главный экран.
type Summary struct {
	Counters
	Money
	Upcoming   []payments.Payment   `json:"upcoming_payments"`
	Overdue    []payments.Payment   `json:"overdue_payments"`
	Recent     []payments.Payment   `json:"recent_operations"`
	EndingSoon []contracts.Contract `json:"contracts_ending_soon_list"`
	Monthly    []MonthPoint         `json:"monthly"`
}
