package ports

import (
	"context"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

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

type MerchantService interface {
	CreateMerchant(ctx context.Context, newMerchant NewMerchant) (pgsqlc.Merchant, *domainerr.DomainError)
	UpdateMerchant(ctx context.Context, merchantID string, newMerchant NewMerchant) (pgsqlc.Merchant, *domainerr.DomainError)
	BootstrapActor(ctx context.Context, merchantID string, actor BootstrapActor) (pgsqlc.CreateActorRow, *domainerr.DomainError)
	GetMerchant(ctx context.Context, merchantID string) (pgsqlc.Merchant, *domainerr.DomainError)
	ListMerchants(ctx context.Context) ([]pgsqlc.Merchant, *domainerr.DomainError)
	ListBranchesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.Branch, *domainerr.DomainError)
	ListDiscountsByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.MerchantDiscount, *domainerr.DomainError)
	ListRolesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.Role, *domainerr.DomainError)
}
