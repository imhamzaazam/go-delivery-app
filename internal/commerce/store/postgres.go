package store

import (
	"context"

	"github.com/google/uuid"
	actorstore "github.com/horiondreher/go-web-api-boilerplate/internal/actor/store"
	cartstore "github.com/horiondreher/go-web-api-boilerplate/internal/cart/store"
	catalogstore "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/store"
	coveragestore "github.com/horiondreher/go-web-api-boilerplate/internal/coverage/store"
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	orderstore "github.com/horiondreher/go-web-api-boilerplate/internal/order/store"
	reportstore "github.com/horiondreher/go-web-api-boilerplate/internal/report/store"

	pgsqlc "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
)

type DBTX = actorstore.DBTX

type Postgres struct {
	actor    *actorstore.Postgres
	catalog  *catalogstore.Postgres
	cart     *cartstore.Postgres
	coverage *coveragestore.Postgres
	merchant *merchantstore.Postgres
	order    *orderstore.Postgres
	report   *reportstore.Postgres
}

func New(db DBTX) *Postgres {
	return &Postgres{
		actor:    actorstore.New(db),
		catalog:  catalogstore.New(db),
		cart:     cartstore.New(db),
		coverage: coveragestore.New(db),
		merchant: merchantstore.New(db),
		order:    orderstore.New(db),
		report:   reportstore.New(db),
	}
}

func (store *Postgres) CreateActor(ctx context.Context, arg pgsqlc.CreateActorParams) (pgsqlc.CreateActorRow, error) {
	return store.actor.CreateActor(ctx, arg)
}

func (store *Postgres) GetActor(ctx context.Context, arg pgsqlc.GetActorParams) (pgsqlc.GetActorRow, error) {
	return store.actor.GetActor(ctx, arg)
}

func (store *Postgres) CreateSession(ctx context.Context, arg pgsqlc.CreateSessionParams) (pgsqlc.CreateSessionRow, error) {
	return store.actor.CreateSession(ctx, arg)
}

func (store *Postgres) GetSession(ctx context.Context, id uuid.UUID) (pgsqlc.GetSessionRow, error) {
	return store.actor.GetSession(ctx, id)
}

func (store *Postgres) GetActorByUID(ctx context.Context, arg pgsqlc.GetActorByUIDParams) (pgsqlc.GetActorByUIDRow, error) {
	return store.actor.GetActorByUID(ctx, arg)
}

func (store *Postgres) GetActorProfileByMerchantAndEmail(ctx context.Context, arg pgsqlc.GetActorProfileByMerchantAndEmailParams) (pgsqlc.GetActorProfileByMerchantAndEmailRow, error) {
	return store.actor.GetActorProfileByMerchantAndEmail(ctx, arg)
}

func (store *Postgres) ListActorsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListActorsByMerchantRow, error) {
	return store.actor.ListActorsByMerchant(ctx, merchantID)
}

func (store *Postgres) ListEmployeesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListEmployeesByMerchantRow, error) {
	return store.actor.ListEmployeesByMerchant(ctx, merchantID)
}

func (store *Postgres) CreateArea(ctx context.Context, arg pgsqlc.CreateAreaParams) (pgsqlc.Area, error) {
	return store.coverage.CreateArea(ctx, arg)
}

func (store *Postgres) GetArea(ctx context.Context, id uuid.UUID) (pgsqlc.Area, error) {
	return store.coverage.GetArea(ctx, id)
}

func (store *Postgres) CreateZone(ctx context.Context, arg pgsqlc.CreateZoneParams) (pgsqlc.CreateZoneRow, error) {
	return store.coverage.CreateZone(ctx, arg)
}

func (store *Postgres) GetZone(ctx context.Context, id uuid.UUID) (pgsqlc.GetZoneRow, error) {
	return store.coverage.GetZone(ctx, id)
}

func (store *Postgres) ListAreas(ctx context.Context) ([]pgsqlc.Area, error) {
	return store.coverage.ListAreas(ctx)
}

func (store *Postgres) ListZonesByArea(ctx context.Context, areaID uuid.UUID) ([]pgsqlc.ListZonesByAreaRow, error) {
	return store.coverage.ListZonesByArea(ctx, areaID)
}

func (store *Postgres) CreateMerchantServiceZone(ctx context.Context, arg pgsqlc.CreateMerchantServiceZoneParams) (pgsqlc.MerchantServiceZone, error) {
	return store.coverage.CreateMerchantServiceZone(ctx, arg)
}

func (store *Postgres) ListMerchantServiceZonesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListMerchantServiceZonesByMerchantRow, error) {
	return store.coverage.ListMerchantServiceZonesByMerchant(ctx, merchantID)
}

func (store *Postgres) CreateMerchant(ctx context.Context, arg pgsqlc.CreateMerchantParams) (pgsqlc.Merchant, error) {
	m, err := store.merchant.CreateMerchant(ctx, arg)
	if err != nil {
		return pgsqlc.Merchant{}, err
	}
	return toMerchant(m), nil
}

func (store *Postgres) UpdateMerchant(ctx context.Context, arg pgsqlc.UpdateMerchantParams) (pgsqlc.Merchant, error) {
	m, err := store.merchant.UpdateMerchant(ctx, arg)
	if err != nil {
		return pgsqlc.Merchant{}, err
	}
	return toMerchant(m), nil
}

func (store *Postgres) GetMerchant(ctx context.Context, id uuid.UUID) (pgsqlc.Merchant, error) {
	m, err := store.merchant.GetMerchant(ctx, id)
	if err != nil {
		return pgsqlc.Merchant{}, err
	}
	return toMerchant(m), nil
}

func (store *Postgres) ListMerchants(ctx context.Context) ([]pgsqlc.Merchant, error) {
	items, err := store.merchant.ListMerchants(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]pgsqlc.Merchant, 0, len(items))
	for _, m := range items {
		result = append(result, toMerchant(m))
	}
	return result, nil
}

func (store *Postgres) CreateRole(ctx context.Context, arg pgsqlc.CreateRoleParams) (pgsqlc.Role, error) {
	r, err := store.merchant.CreateRole(ctx, arg)
	if err != nil {
		return pgsqlc.Role{}, err
	}
	return toRole(r), nil
}

func (store *Postgres) ListRolesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.Role, error) {
	items, err := store.merchant.ListRolesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]pgsqlc.Role, 0, len(items))
	for _, r := range items {
		result = append(result, toRole(r))
	}
	return result, nil
}

func (store *Postgres) CreateBranch(ctx context.Context, arg pgsqlc.CreateBranchParams) (pgsqlc.Branch, error) {
	b, err := store.merchant.CreateBranch(ctx, arg)
	if err != nil {
		return pgsqlc.Branch{}, err
	}
	return toBranch(b), nil
}

func (store *Postgres) GetBranch(ctx context.Context, arg pgsqlc.GetBranchParams) (pgsqlc.Branch, error) {
	b, err := store.merchant.GetBranch(ctx, arg)
	if err != nil {
		return pgsqlc.Branch{}, err
	}
	return toBranch(b), nil
}

func (store *Postgres) ListBranchesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListBranchesByMerchantRow, error) {
	return store.merchant.ListBranchesByMerchant(ctx, merchantID)
}

func (store *Postgres) CreateMerchantDiscount(ctx context.Context, arg pgsqlc.CreateMerchantDiscountParams) (pgsqlc.CreateMerchantDiscountRow, error) {
	return store.merchant.CreateMerchantDiscount(ctx, arg)
}

func (store *Postgres) GetMerchantDiscount(ctx context.Context, arg pgsqlc.GetMerchantDiscountParams) (pgsqlc.GetMerchantDiscountRow, error) {
	return store.merchant.GetMerchantDiscount(ctx, arg)
}

func (store *Postgres) ListDiscountsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListDiscountsByMerchantRow, error) {
	return store.merchant.ListDiscountsByMerchant(ctx, merchantID)
}

func (store *Postgres) CreateProductCategory(ctx context.Context, arg pgsqlc.CreateProductCategoryParams) (pgsqlc.ProductCategory, error) {
	return store.catalog.CreateProductCategory(ctx, arg)
}

func (store *Postgres) GetProductCategory(ctx context.Context, arg pgsqlc.GetProductCategoryParams) (pgsqlc.ProductCategory, error) {
	return store.catalog.GetProductCategory(ctx, arg)
}

func (store *Postgres) CreateProduct(ctx context.Context, arg pgsqlc.CreateProductParams) (pgsqlc.Product, error) {
	return store.catalog.CreateProduct(ctx, arg)
}

func (store *Postgres) GetProduct(ctx context.Context, arg pgsqlc.GetProductParams) (pgsqlc.Product, error) {
	return store.catalog.GetProduct(ctx, arg)
}

func (store *Postgres) GetProductAddon(ctx context.Context, id uuid.UUID) (pgsqlc.ProductAddon, error) {
	return store.catalog.GetProductAddon(ctx, id)
}

func (store *Postgres) CreateProductAddon(ctx context.Context, arg pgsqlc.CreateProductAddonParams) (pgsqlc.ProductAddon, error) {
	return store.catalog.CreateProductAddon(ctx, arg)
}

func (store *Postgres) ListProductCategoriesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ProductCategory, error) {
	return store.catalog.ListProductCategoriesByMerchant(ctx, merchantID)
}

func (store *Postgres) ListProductsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.Product, error) {
	return store.catalog.ListProductsByMerchant(ctx, merchantID)
}

func (store *Postgres) GetProductDetail(ctx context.Context, arg pgsqlc.GetProductDetailParams) (pgsqlc.GetProductDetailRow, error) {
	return store.catalog.GetProductDetail(ctx, arg)
}

func (store *Postgres) ListProductAddonsByProduct(ctx context.Context, productID uuid.UUID) ([]pgsqlc.ProductAddon, error) {
	return store.catalog.ListProductAddonsByProduct(ctx, productID)
}

func (store *Postgres) ListInventoryByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListInventoryByMerchantRow, error) {
	return store.catalog.ListInventoryByMerchant(ctx, merchantID)
}

func (store *Postgres) GetProductInventory(ctx context.Context, arg pgsqlc.GetProductInventoryParams) (pgsqlc.ProductInventory, error) {
	return store.catalog.GetProductInventory(ctx, arg)
}

func (store *Postgres) UpdateProductInventoryQuantity(ctx context.Context, arg pgsqlc.UpdateProductInventoryQuantityParams) (pgsqlc.ProductInventory, error) {
	return store.catalog.UpdateProductInventoryQuantity(ctx, arg)
}

func (store *Postgres) UpsertProductInventory(ctx context.Context, arg pgsqlc.UpsertProductInventoryParams) (pgsqlc.ProductInventory, error) {
	return store.catalog.UpsertProductInventory(ctx, arg)
}

func (store *Postgres) CreateCart(ctx context.Context, arg pgsqlc.CreateCartParams) (pgsqlc.Cart, error) {
	return store.cart.CreateCart(ctx, arg)
}

func (store *Postgres) CreateGuestCart(ctx context.Context, arg pgsqlc.CreateGuestCartParams) (pgsqlc.Cart, error) {
	return store.cart.CreateGuestCart(ctx, arg)
}

func (store *Postgres) GetCart(ctx context.Context, arg pgsqlc.GetCartParams) (pgsqlc.Cart, error) {
	return store.cart.GetCart(ctx, arg)
}

func (store *Postgres) UpdateCart(ctx context.Context, arg pgsqlc.UpdateCartParams) (pgsqlc.Cart, error) {
	return store.cart.UpdateCart(ctx, arg)
}

func (store *Postgres) CreateCartItem(ctx context.Context, arg pgsqlc.CreateCartItemParams) (pgsqlc.CartItem, error) {
	return store.cart.CreateCartItem(ctx, arg)
}

func (store *Postgres) GetCartItemBySignature(ctx context.Context, arg pgsqlc.GetCartItemBySignatureParams) (pgsqlc.CartItem, error) {
	return store.cart.GetCartItemBySignature(ctx, arg)
}

func (store *Postgres) GetCartItemByID(ctx context.Context, arg pgsqlc.GetCartItemByIDParams) (pgsqlc.CartItem, error) {
	return store.cart.GetCartItemByID(ctx, arg)
}

func (store *Postgres) UpdateCartItemByID(ctx context.Context, arg pgsqlc.UpdateCartItemByIDParams) (pgsqlc.CartItem, error) {
	return store.cart.UpdateCartItemByID(ctx, arg)
}

func (store *Postgres) DeleteCartItem(ctx context.Context, arg pgsqlc.DeleteCartItemParams) (int64, error) {
	return store.cart.DeleteCartItem(ctx, arg)
}

func (store *Postgres) ListCartItemsByCart(ctx context.Context, cartID uuid.UUID) ([]pgsqlc.CartItem, error) {
	return store.cart.ListCartItemsByCart(ctx, cartID)
}

func (store *Postgres) CreateOrder(ctx context.Context, arg pgsqlc.CreateOrderParams) (pgsqlc.Order, error) {
	return store.order.CreateOrder(ctx, arg)
}

func (store *Postgres) CreateOrderGuest(ctx context.Context, arg pgsqlc.CreateOrderGuestParams) (pgsqlc.Order, error) {
	return store.order.CreateOrderGuest(ctx, arg)
}

func (store *Postgres) GetOrder(ctx context.Context, arg pgsqlc.GetOrderParams) (pgsqlc.Order, error) {
	return store.order.GetOrder(ctx, arg)
}

func (store *Postgres) UpdateOrder(ctx context.Context, arg pgsqlc.UpdateOrderParams) (pgsqlc.Order, error) {
	return store.order.UpdateOrder(ctx, arg)
}

func (store *Postgres) UpsertOrderItem(ctx context.Context, arg pgsqlc.UpsertOrderItemParams) (pgsqlc.OrderItem, error) {
	return store.order.UpsertOrderItem(ctx, arg)
}

func (store *Postgres) GetOrderItem(ctx context.Context, arg pgsqlc.GetOrderItemParams) (pgsqlc.OrderItem, error) {
	return store.order.GetOrderItem(ctx, arg)
}

func (store *Postgres) UpdateOrderItem(ctx context.Context, arg pgsqlc.UpdateOrderItemParams) (pgsqlc.OrderItem, error) {
	return store.order.UpdateOrderItem(ctx, arg)
}

func (store *Postgres) UpsertOrderItemAddon(ctx context.Context, arg pgsqlc.UpsertOrderItemAddonParams) (pgsqlc.OrderItemAddon, error) {
	return store.order.UpsertOrderItemAddon(ctx, arg)
}

func (store *Postgres) GetOrderItemAddon(ctx context.Context, arg pgsqlc.GetOrderItemAddonParams) (pgsqlc.OrderItemAddon, error) {
	return store.order.GetOrderItemAddon(ctx, arg)
}

func (store *Postgres) UpdateOrderItemAddon(ctx context.Context, arg pgsqlc.UpdateOrderItemAddonParams) (pgsqlc.OrderItemAddon, error) {
	return store.order.UpdateOrderItemAddon(ctx, arg)
}

func (store *Postgres) ListOrdersByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.Order, error) {
	return store.order.ListOrdersByMerchant(ctx, merchantID)
}

func (store *Postgres) ListOrderItemsByOrder(ctx context.Context, orderID uuid.UUID) ([]pgsqlc.OrderItem, error) {
	return store.order.ListOrderItemsByOrder(ctx, orderID)
}

func (store *Postgres) ListOrderItemAddonsByOrder(ctx context.Context, orderID uuid.UUID) ([]pgsqlc.OrderItemAddon, error) {
	return store.order.ListOrderItemAddonsByOrder(ctx, orderID)
}

func (store *Postgres) GetVatRule(ctx context.Context, arg pgsqlc.GetVatRuleParams) (pgsqlc.VatRule, error) {
	return store.order.GetVatRule(ctx, arg)
}

func (store *Postgres) UpdateVatRuleByID(ctx context.Context, arg pgsqlc.UpdateVatRuleByIDParams) (pgsqlc.VatRule, error) {
	return store.order.UpdateVatRuleByID(ctx, arg)
}

func (store *Postgres) UpsertVatRule(ctx context.Context, arg pgsqlc.UpsertVatRuleParams) (pgsqlc.VatRule, error) {
	return store.order.UpsertVatRule(ctx, arg)
}

func (store *Postgres) GetMonthlySalesReport(ctx context.Context, arg pgsqlc.GetMonthlySalesReportParams) (pgsqlc.GetMonthlySalesReportRow, error) {
	return store.report.GetMonthlySalesReport(ctx, arg)
}
