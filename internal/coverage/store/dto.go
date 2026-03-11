package store

import (
	"github.com/horiondreher/go-web-api-boilerplate/internal/coverage/store/generated"
)

type Area = coveragestore.Area
type CityType = coveragestore.CityType
type CreateAreaParams = coveragestore.CreateAreaParams
type CreateZoneParams = coveragestore.CreateZoneParams
type CreateZoneRow = coveragestore.CreateZoneRow
type GetZoneRow = coveragestore.GetZoneRow
type ListMerchantServiceZonesByMerchantRow = coveragestore.ListMerchantServiceZonesByMerchantRow
type ListZonesByAreaRow = coveragestore.ListZonesByAreaRow
type MerchantServiceZone = coveragestore.MerchantServiceZone
type CreateMerchantServiceZoneParams = coveragestore.CreateMerchantServiceZoneParams

const CityTypeKarachi = coveragestore.CityTypeKarachi
const CityTypeLahore = coveragestore.CityTypeLahore
