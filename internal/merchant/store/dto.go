package store

import (
	"github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store/generated"
)

type Branch = merchantstore.Branch
type CreateBranchParams = merchantstore.CreateBranchParams
type CreateBranchRow = merchantstore.Branch
type GetBranchParams = merchantstore.GetBranchParams
type GetBranchRow = merchantstore.Branch
type ListBranchesByMerchantRow = merchantstore.Branch
type Merchant = merchantstore.Merchant
type MerchantCategory = merchantstore.MerchantCategory
type MerchantDiscount = merchantstore.ListDiscountsByMerchantRow
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

const CityTypeKarachi = merchantstore.CityTypeKarachi
const CityTypeLahore = merchantstore.CityTypeLahore
const DiscountTypeFlat = merchantstore.DiscountTypeFlat
const DiscountTypePercentage = merchantstore.DiscountTypePercentage
const RoleTypeAdmin = merchantstore.RoleTypeAdmin
const RoleTypeCustomer = merchantstore.RoleTypeCustomer
const RoleTypeEmployee = merchantstore.RoleTypeEmployee
const RoleTypeMerchant = merchantstore.RoleTypeMerchant
