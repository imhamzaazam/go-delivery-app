package merchant

import (
	"context"

	actor "github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	coveragestore "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/store"
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
)

type Branch = merchantstore.Branch
type CreateBranchParams = merchantstore.CreateBranchParams
type CreateBranchRow = merchantstore.CreateBranchRow
type GetBranchParams = merchantstore.GetBranchParams
type GetBranchRow = merchantstore.GetBranchRow
type ListBranchesByMerchantRow = merchantstore.ListBranchesByMerchantRow
type Merchant = merchantstore.Merchant
type MerchantCategory = merchantstore.MerchantCategory
type MerchantDiscount = merchantstore.MerchantDiscount
type CreateMerchantParams = merchantstore.CreateMerchantParams
type UpdateMerchantParams = merchantstore.UpdateMerchantParams
type CreateMerchantDiscountParams = merchantstore.CreateMerchantDiscountParams
type CreateMerchantDiscountRow = merchantstore.CreateMerchantDiscountRow
type GetMerchantDiscountParams = merchantstore.GetMerchantDiscountParams
type GetMerchantDiscountRow = merchantstore.GetMerchantDiscountRow
type ListDiscountsByMerchantRow = merchantstore.ListDiscountsByMerchantRow
type Role = merchantstore.Role
type CreateRoleParams = merchantstore.CreateRoleParams
type CityType = merchantstore.CityType
type DiscountType = merchantstore.DiscountType
type RoleType = merchantstore.RoleType
type MerchantServiceZone = coveragestore.MerchantServiceZone

const CityTypeKarachi = merchantstore.CityTypeKarachi
const CityTypeLahore = merchantstore.CityTypeLahore

const DiscountTypeFlat = merchantstore.DiscountTypeFlat
const DiscountTypePercentage = merchantstore.DiscountTypePercentage

const RoleTypeAdmin = merchantstore.RoleTypeAdmin
const RoleTypeCustomer = merchantstore.RoleTypeCustomer
const RoleTypeEmployee = merchantstore.RoleTypeEmployee
const RoleTypeMerchant = merchantstore.RoleTypeMerchant

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

type Service interface {
	CreateMerchant(ctx context.Context, newMerchant NewMerchant) (Merchant, *domainerr.DomainError)
	UpdateMerchant(ctx context.Context, merchantID string, newMerchant NewMerchant) (Merchant, *domainerr.DomainError)
	BootstrapActor(ctx context.Context, merchantID string, actor BootstrapActor) (actor.CreateActorRow, *domainerr.DomainError)
	GetMerchant(ctx context.Context, merchantID string) (Merchant, *domainerr.DomainError)
	ListMerchants(ctx context.Context) ([]Merchant, *domainerr.DomainError)
	ListBranchesByMerchant(ctx context.Context, merchantID string) ([]Branch, *domainerr.DomainError)
	ListDiscountsByMerchant(ctx context.Context, merchantID string) ([]MerchantDiscount, *domainerr.DomainError)
	ListRolesByMerchant(ctx context.Context, merchantID string) ([]Role, *domainerr.DomainError)
}
