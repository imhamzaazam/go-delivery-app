package v1

import (
	"net/http"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
)

var _ api.ServerInterface = (*HTTPAdapter)(nil)

func (adapter *HTTPAdapter) GetActorByUIDLegacy(w http.ResponseWriter, r *http.Request, uid string) {
	adapter.actorHandler.GetActorByUIDLegacy(w, r, uid)
}

func (adapter *HTTPAdapter) CreateActor(w http.ResponseWriter, r *http.Request) {
	adapter.actorHandler.CreateActor(w, r)
}

func (adapter *HTTPAdapter) GetAuthenticatedActor(w http.ResponseWriter, r *http.Request) {
	adapter.actorHandler.GetAuthenticatedActor(w, r)
}

func (adapter *HTTPAdapter) GetActorByUID(w http.ResponseWriter, r *http.Request, uid string) {
	adapter.actorHandler.GetActorByUID(w, r, uid)
}

func (adapter *HTTPAdapter) ListAreas(w http.ResponseWriter, r *http.Request) {
	adapter.coverageHandler.ListAreas(w, r)
}

func (adapter *HTTPAdapter) CreateArea(w http.ResponseWriter, r *http.Request) {
	adapter.coverageHandler.CreateArea(w, r)
}

func (adapter *HTTPAdapter) ListZonesByArea(w http.ResponseWriter, r *http.Request, areaId string) {
	adapter.coverageHandler.ListZonesByArea(w, r, areaId)
}

func (adapter *HTTPAdapter) CreateZone(w http.ResponseWriter, r *http.Request, areaId string) {
	adapter.coverageHandler.CreateZone(w, r, areaId)
}

func (adapter *HTTPAdapter) CreateCart(w http.ResponseWriter, r *http.Request) {
	adapter.cartHandler.CreateCart(w, r)
}

func (adapter *HTTPAdapter) GetCartDetail(w http.ResponseWriter, r *http.Request, cartId string, params api.GetCartDetailParams) {
	adapter.cartHandler.GetCartDetail(w, r, cartId, params)
}

func (adapter *HTTPAdapter) AddItemToCart(w http.ResponseWriter, r *http.Request, cartId string) {
	adapter.cartHandler.AddItemToCart(w, r, cartId)
}

func (adapter *HTTPAdapter) DeleteCartItem(w http.ResponseWriter, r *http.Request, cartId string, itemId string) {
	adapter.cartHandler.DeleteCartItem(w, r, cartId, itemId)
}

func (adapter *HTTPAdapter) UpdateCartItem(w http.ResponseWriter, r *http.Request, cartId string, itemId string) {
	adapter.cartHandler.UpdateCartItem(w, r, cartId, itemId)
}

func (adapter *HTTPAdapter) LoginActor(w http.ResponseWriter, r *http.Request) {
	adapter.authHandler.LoginActor(w, r)
}

func (adapter *HTTPAdapter) GetMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.GetMerchant(w, r)
}

func (adapter *HTTPAdapter) UpdateMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.UpdateMerchant(w, r)
}

func (adapter *HTTPAdapter) ListActorsByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.actorHandler.ListActorsByMerchant(w, r)
}

func (adapter *HTTPAdapter) ListBranchesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.ListBranchesByMerchant(w, r)
}

func (adapter *HTTPAdapter) CreateBranchByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.CreateBranchByMerchant(w, r)
}

func (adapter *HTTPAdapter) GetBranchAvailability(w http.ResponseWriter, r *http.Request, branchId string) {
	adapter.merchantHandler.GetBranchAvailability(w, r, branchId)
}

func (adapter *HTTPAdapter) ListProductCategoriesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.catalogHandler.ListProductCategoriesByMerchant(w, r)
}

func (adapter *HTTPAdapter) CreateProductCategoryByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.catalogHandler.CreateProductCategoryByMerchant(w, r)
}

func (adapter *HTTPAdapter) ListDiscountsByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.ListDiscountsByMerchant(w, r)
}

func (adapter *HTTPAdapter) CreateDiscountByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.CreateDiscountByMerchant(w, r)
}

func (adapter *HTTPAdapter) ListEmployeesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.actorHandler.ListEmployeesByMerchant(w, r)
}

func (adapter *HTTPAdapter) ListInventoryByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.catalogHandler.ListInventoryByMerchant(w, r)
}

func (adapter *HTTPAdapter) UpsertInventoryByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.catalogHandler.UpsertInventoryByMerchant(w, r)
}

func (adapter *HTTPAdapter) ListOrdersByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.orderHandler.ListOrdersByMerchant(w, r)
}

func (adapter *HTTPAdapter) ListProductsByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.catalogHandler.ListProductsByMerchant(w, r)
}

func (adapter *HTTPAdapter) CreateProductByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.catalogHandler.CreateProductByMerchant(w, r)
}

func (adapter *HTTPAdapter) GetMonthlySalesReport(w http.ResponseWriter, r *http.Request, params api.GetMonthlySalesReportParams) {
	adapter.reportHandler.GetMonthlySalesReport(w, r, params)
}

func (adapter *HTTPAdapter) ListRolesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.ListRolesByMerchant(w, r)
}

func (adapter *HTTPAdapter) ListMerchantServiceZonesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.coverageHandler.ListMerchantServiceZonesByMerchant(w, r)
}

func (adapter *HTTPAdapter) CreateMerchantServiceZoneByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.coverageHandler.CreateMerchantServiceZoneByMerchant(w, r)
}

func (adapter *HTTPAdapter) CheckMerchantServiceZoneCoverage(w http.ResponseWriter, r *http.Request) {
	adapter.coverageHandler.CheckMerchantServiceZoneCoverage(w, r)
}

func (adapter *HTTPAdapter) ListMerchants(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.ListMerchants(w, r)
}

func (adapter *HTTPAdapter) CreateMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.merchantHandler.CreateMerchant(w, r)
}

func (adapter *HTTPAdapter) BootstrapMerchantActor(w http.ResponseWriter, r *http.Request, merchantId string) {
	adapter.merchantHandler.BootstrapMerchantActor(w, r, merchantId)
}

func (adapter *HTTPAdapter) PlaceOrderFromCart(w http.ResponseWriter, r *http.Request) {
	adapter.orderHandler.PlaceOrderFromCart(w, r)
}

func (adapter *HTTPAdapter) GetOrderDetail(w http.ResponseWriter, r *http.Request, orderId string) {
	adapter.orderHandler.GetOrderDetail(w, r, orderId)
}

func (adapter *HTTPAdapter) UpdateOrderStatus(w http.ResponseWriter, r *http.Request, orderId string) {
	adapter.orderHandler.UpdateOrderStatus(w, r, orderId)
}

func (adapter *HTTPAdapter) GetProductDetail(w http.ResponseWriter, r *http.Request, productId string) {
	adapter.catalogHandler.GetProductDetail(w, r, productId)
}

func (adapter *HTTPAdapter) ListProductAddonsByProduct(w http.ResponseWriter, r *http.Request, productId string) {
	adapter.catalogHandler.ListProductAddonsByProduct(w, r, productId)
}

func (adapter *HTTPAdapter) AddProductAddonByMerchant(w http.ResponseWriter, r *http.Request, productId string) {
	adapter.catalogHandler.AddProductAddonByMerchant(w, r, productId)
}

func (adapter *HTTPAdapter) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	adapter.authHandler.RenewAccessToken(w, r)
}
