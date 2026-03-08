package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type ReadService interface {
	ListProductCategoriesByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.ProductCategory, *domainerr.DomainError)
	ListProductsByMerchant(ctx context.Context, merchantID string) ([]pgsqlc.Product, *domainerr.DomainError)
	GetProductDetail(ctx context.Context, merchantID string, productID string) (pgsqlc.GetProductDetailRow, *domainerr.DomainError)
	ListProductAddonsByProduct(ctx context.Context, merchantID string, productID string) ([]pgsqlc.ProductAddon, *domainerr.DomainError)
	ListInventoryByMerchant(ctx context.Context, viewerActorID uuid.UUID, merchantID string) ([]pgsqlc.ListInventoryByMerchantRow, *domainerr.DomainError)
	GetCartDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, cartID string, paymentType string) (CartDetail, *domainerr.DomainError)
	GetOrderDetail(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, orderID string) (OrderDetail, *domainerr.DomainError)
	ListOrdersByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]pgsqlc.Order, *domainerr.DomainError)
	GetMonthlySalesReport(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, month int, year int) (SalesReport, *domainerr.DomainError)
	ListAreas(ctx context.Context) ([]pgsqlc.Area, *domainerr.DomainError)
	ListZonesByArea(ctx context.Context, areaID string) ([]pgsqlc.ListZonesByAreaRow, *domainerr.DomainError)
	ListMerchantServiceZonesByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]pgsqlc.ListMerchantServiceZonesByMerchantRow, *domainerr.DomainError)
	CheckMerchantServiceZoneCoverage(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, latitude float64, longitude float64) (ServiceZoneCoverageResult, *domainerr.DomainError)
}

type ServiceZoneCoverageResult struct {
	Covered    bool
	MerchantID uuid.UUID
	ZoneID     uuid.UUID
	ZoneName   string
	BranchID   uuid.UUID
	BranchName string
	AreaID     uuid.UUID
	AreaName   string
	AreaCity   string
}
