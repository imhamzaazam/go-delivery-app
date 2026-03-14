package merchant

import (
	"time"

	"github.com/google/uuid"
)

type MerchantCategory string

const (
	MerchantCategoryRestaurant MerchantCategory = "restaurant"
	MerchantCategoryPharma     MerchantCategory = "pharma"
	MerchantCategoryBakery     MerchantCategory = "bakery"
)

type CityType string

const (
	CityTypeKarachi CityType = "Karachi"
	CityTypeLahore  CityType = "Lahore"
)

type RoleType string

const (
	RoleTypeAdmin    RoleType = "admin"
	RoleTypeMerchant RoleType = "merchant"
	RoleTypeEmployee RoleType = "employee"
	RoleTypeCustomer RoleType = "customer"
)

type DiscountType string

const (
	DiscountTypeFlat       DiscountType = "flat"
	DiscountTypePercentage DiscountType = "percentage"
)

type Merchant struct {
	ID            uuid.UUID
	Name          string
	Ntn           string
	Address       string
	Logo          *string
	Category      MerchantCategory
	ContactNumber string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Branch struct {
	ID                 uuid.UUID
	MerchantID         uuid.UUID
	Name               string
	Address            string
	ContactNumber      *string
	City               CityType
	OpeningTimeMinutes int16
	ClosingTimeMinutes int16
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type MerchantDiscount struct {
	ID          uuid.UUID
	MerchantID  uuid.UUID
	Type        DiscountType
	Value       float64
	Description *string
	ValidFrom   time.Time
	ValidTo     time.Time
	CreatedAt   time.Time
	ProductID   uuid.UUID
	CategoryID  uuid.UUID
}

type Role struct {
	ID          uuid.UUID
	MerchantID  uuid.UUID
	RoleType    RoleType
	Description *string
	CreatedAt   time.Time
}

type NewMerchant struct {
	Name          string
	Ntn           string
	Address       string
	Category      string
	ContactNumber string
}

type BootstrapActor struct {
	FullName string
	Email    string
	Password string
	Role     string
}

type BranchAvailability struct {
	Branch       Branch
	IsOpen       bool
	OpeningTime  string
	ClosingTime  string
	CurrentTime  string
	TimezoneName string
}
