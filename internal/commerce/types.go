package commerce

import (
	"github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	"github.com/horiondreher/go-web-api-boilerplate/internal/auth"
	"github.com/horiondreher/go-web-api-boilerplate/internal/cart"
	"github.com/horiondreher/go-web-api-boilerplate/internal/catalog"
	"github.com/horiondreher/go-web-api-boilerplate/internal/coverage"
	"github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	"github.com/horiondreher/go-web-api-boilerplate/internal/order"
	"github.com/horiondreher/go-web-api-boilerplate/internal/report"
)

type Actor = actor.Actor
type CreateActorParams = actor.CreateActorParams
type CreateActorRow = actor.CreateActorRow
type GetActorParams = actor.GetActorParams
type GetActorRow = actor.GetActorRow
type GetActorByUIDParams = actor.GetActorByUIDParams
type GetActorByUIDRow = actor.GetActorByUIDRow
type GetActorProfileByMerchantAndEmailParams = actor.GetActorProfileByMerchantAndEmailParams
type GetActorProfileByMerchantAndEmailRow = actor.GetActorProfileByMerchantAndEmailRow
type ListActorsByMerchantRow = actor.ListActorsByMerchantRow
type ListEmployeesByMerchantRow = actor.ListEmployeesByMerchantRow
type CreateSessionParams = actor.CreateSessionParams
type CreateSessionRow = actor.CreateSessionRow
type GetSessionRow = actor.GetSessionRow

type DBTX = auth.DBTX
type ActorRole = auth.ActorRole
type AssignActorRoleParams = auth.AssignActorRoleParams
type GetActorRoleParams = auth.GetActorRoleParams
type TouchActorRoleAssignedAtParams = auth.TouchActorRoleAssignedAtParams

type Cart = cart.Cart
type CartItem = cart.CartItem
type CreateCartParams = cart.CreateCartParams
type CreateGuestCartParams = cart.CreateGuestCartParams
type GetCartParams = cart.GetCartParams
type UpdateCartParams = cart.UpdateCartParams
type CreateCartItemParams = cart.CreateCartItemParams
type CreateCartItemRow = cart.CreateCartItemRow
type GetCartItemBySignatureParams = cart.GetCartItemBySignatureParams
type GetCartItemBySignatureRow = cart.GetCartItemBySignatureRow
type GetCartItemByIDParams = cart.GetCartItemByIDParams
type GetCartItemByIDRow = cart.GetCartItemByIDRow
type UpdateCartItemByIDParams = cart.UpdateCartItemByIDParams
type UpdateCartItemByIDRow = cart.UpdateCartItemByIDRow
type DeleteCartItemParams = cart.DeleteCartItemParams
type ListCartItemsByCartRow = cart.ListCartItemsByCartRow

type Product = catalog.Product
type ProductAddon = catalog.ProductAddon
type ProductCategory = catalog.ProductCategory
type ProductInventory = catalog.ProductInventory
type CreateProductParams = catalog.CreateProductParams
type CreateProductAddonParams = catalog.CreateProductAddonParams
type CreateProductCategoryParams = catalog.CreateProductCategoryParams
type GetProductParams = catalog.GetProductParams
type GetProductCategoryParams = catalog.GetProductCategoryParams
type GetProductDetailParams = catalog.GetProductDetailParams
type GetProductDetailRow = catalog.GetProductDetailRow
type GetProductInventoryParams = catalog.GetProductInventoryParams
type ListInventoryByMerchantRow = catalog.ListInventoryByMerchantRow
type UpdateProductInventoryQuantityParams = catalog.UpdateProductInventoryQuantityParams
type UpsertProductInventoryParams = catalog.UpsertProductInventoryParams

type Area = coverage.Area
type CreateAreaParams = coverage.CreateAreaParams
type CreateZoneParams = coverage.CreateZoneParams
type CreateZoneRow = coverage.CreateZoneRow
type GetZoneRow = coverage.GetZoneRow
type ListMerchantServiceZonesByMerchantRow = coverage.ListMerchantServiceZonesByMerchantRow
type ListZonesByAreaRow = coverage.ListZonesByAreaRow
type MerchantServiceZone = coverage.MerchantServiceZone
type CreateMerchantServiceZoneParams = coverage.CreateMerchantServiceZoneParams

type Branch = merchant.Branch
type CreateBranchParams = merchantstore.CreateBranchParams
type CreateBranchRow = merchantstore.CreateBranchRow
type GetBranchParams = merchantstore.GetBranchParams
type GetBranchRow = merchantstore.GetBranchRow
type ListBranchesByMerchantRow = merchantstore.ListBranchesByMerchantRow
type Merchant = merchant.Merchant
type MerchantCategory = merchant.MerchantCategory
type MerchantDiscount = merchant.MerchantDiscount
type CreateMerchantParams = merchantstore.CreateMerchantParams
type UpdateMerchantParams = merchantstore.UpdateMerchantParams
type CreateMerchantDiscountParams = merchantstore.CreateMerchantDiscountParams
type CreateMerchantDiscountRow = merchantstore.CreateMerchantDiscountRow
type GetMerchantDiscountParams = merchantstore.GetMerchantDiscountParams
type GetMerchantDiscountRow = merchantstore.GetMerchantDiscountRow
type ListDiscountsByMerchantRow = merchantstore.ListDiscountsByMerchantRow
type Role = merchant.Role
type CreateRoleParams = merchantstore.CreateRoleParams
type CityType = merchant.CityType
type DiscountType = merchant.DiscountType
type RoleType = merchant.RoleType

const CityTypeKarachi = merchant.CityTypeKarachi
const CityTypeLahore = merchant.CityTypeLahore

const DiscountTypeFlat = merchant.DiscountTypeFlat
const DiscountTypePercentage = merchant.DiscountTypePercentage

const RoleTypeAdmin = merchant.RoleTypeAdmin
const RoleTypeCustomer = merchant.RoleTypeCustomer
const RoleTypeEmployee = merchant.RoleTypeEmployee
const RoleTypeMerchant = merchant.RoleTypeMerchant

type Order = order.Order
type OrderItem = order.OrderItem
type OrderItemAddon = order.OrderItemAddon
type OrderStatusType = order.OrderStatusType
type PaymentType = order.PaymentType
type VatRule = order.VatRule
type CreateOrderParams = order.CreateOrderParams
type CreateOrderGuestParams = order.CreateOrderGuestParams
type GetOrderParams = order.GetOrderParams
type UpdateOrderParams = order.UpdateOrderParams
type UpsertOrderItemParams = order.UpsertOrderItemParams
type GetOrderItemParams = order.GetOrderItemParams
type UpdateOrderItemParams = order.UpdateOrderItemParams
type UpsertOrderItemAddonParams = order.UpsertOrderItemAddonParams
type GetOrderItemAddonParams = order.GetOrderItemAddonParams
type UpdateOrderItemAddonParams = order.UpdateOrderItemAddonParams
type GetVatRuleParams = order.GetVatRuleParams
type UpdateVatRuleByIDParams = order.UpdateVatRuleByIDParams
type UpsertVatRuleParams = order.UpsertVatRuleParams

const OrderStatusTypeAccepted = order.OrderStatusTypeAccepted
const OrderStatusTypeCancelled = order.OrderStatusTypeCancelled
const OrderStatusTypeDelivered = order.OrderStatusTypeDelivered
const OrderStatusTypeOutForDelivery = order.OrderStatusTypeOutForDelivery
const OrderStatusTypePending = order.OrderStatusTypePending
const OrderStatusTypeRefunded = order.OrderStatusTypeRefunded

const PaymentTypeCard = order.PaymentTypeCard
const PaymentTypeCash = order.PaymentTypeCash

type GetMonthlySalesReportParams = report.GetMonthlySalesReportParams
type GetMonthlySalesReportRow = report.GetMonthlySalesReportRow
