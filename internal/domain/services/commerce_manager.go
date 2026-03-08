package services

import (
	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderLineBill struct {
	ProductID      uuid.UUID
	Quantity       int32
	BaseAmount     float64
	AddonAmount    float64
	DiscountAmount float64
	TaxAmount      float64
	LineTotal      float64
}

type OrderBill struct {
	OrderID   uuid.UUID
	VatRate   float64
	Subtotal  float64
	TotalTax  float64
	Total     float64
	LineItems []OrderLineBill
}

type SalesReport struct {
	Month          int
	Year           int
	TotalSales     float64
	TotalTax       float64
	TotalDiscount  float64
	ProfitEstimate float64
}

type InventoryItem struct {
	ProductID   uuid.UUID
	ProductName string
	Quantity    int32
}

type CommerceManager struct {
	db    *pgxpool.Pool
	store pgsqlc.Querier
}

func NewCommerceManager(db *pgxpool.Pool, store pgsqlc.Querier) *CommerceManager {
	return &CommerceManager{db: db, store: store}
}
