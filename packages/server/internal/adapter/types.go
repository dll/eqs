package adapter

import "time"

const (
	ProjectID      = "eqs"
	Source         = "eqs"
	SchemaVersion  = "v1"
	AdapterVersion = "1.0.0"
)

// Metadata 是所有适配器响应都必须携带的稳定元数据。
type Metadata struct {
	ProjectID     string    `json:"project_id"`
	Source        string    `json:"source"`
	ObservedAt    time.Time `json:"observed_at"`
	SchemaVersion string    `json:"schema_version"`
}

type response struct {
	Metadata
	Status         string `json:"status,omitempty"`
	Message        string `json:"message,omitempty"`
	Error          string `json:"error,omitempty"`
	Data           any    `json:"data,omitempty"`
	AdapterVersion string `json:"adapter_version,omitempty"`
	EQSVersion     string `json:"eqs_version,omitempty"`
}

type pageData[T any] struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Items    []T   `json:"items"`
}

type projectItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	BudgetRange string    `json:"budget_range"`
	CreatedAt   time.Time `json:"created_at"`
	OwnerID     string    `json:"owner_id"`
	OrderCount  int64     `json:"order_count"`
}

type orderItem struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	AmountRange     string     `json:"amount_range"`
	ServiceProvider string     `json:"service_provider"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type projectSummary struct {
	ID               string      `json:"id"`
	Title            string      `json:"title"`
	Type             string      `json:"type"`
	Status           string      `json:"status"`
	Description      string      `json:"description"`
	BudgetRange      string      `json:"budget_range"`
	Address          string      `json:"address"`
	CreatedAt        time.Time   `json:"created_at"`
	Orders           []orderItem `json:"orders"`
	AttachmentsCount int64       `json:"attachments_count"`
}

type alertItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	ProjectID string    `json:"project_id"`
	OrderID   string    `json:"order_id,omitempty"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
