package store

import (
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

func toMerchant(m merchantstore.Merchant) merchant.Merchant {
	var logo *string
	if m.Logo.Valid {
		logo = &m.Logo.String
	}
	return merchant.Merchant{
		ID:            m.ID,
		Name:          m.Name,
		Ntn:           m.Ntn,
		Address:       m.Address,
		Logo:          logo,
		Category:      merchant.MerchantCategory(m.Category),
		ContactNumber: m.ContactNumber,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func toBranch(b merchantstore.Branch) merchant.Branch {
	var contact *string
	if b.ContactNumber.Valid {
		contact = &b.ContactNumber.String
	}
	return merchant.Branch{
		ID:                 b.ID,
		MerchantID:         b.MerchantID,
		Name:               b.Name,
		Address:            b.Address,
		ContactNumber:      contact,
		City:               merchant.CityType(b.City),
		OpeningTimeMinutes: b.OpeningTimeMinutes,
		ClosingTimeMinutes: b.ClosingTimeMinutes,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}
}

func toRole(r merchantstore.Role) merchant.Role {
	var desc *string
	if r.Description.Valid {
		desc = &r.Description.String
	}
	return merchant.Role{
		ID:          r.ID,
		MerchantID:  r.MerchantID,
		RoleType:    merchant.RoleType(r.RoleType),
		Description: desc,
		CreatedAt:   r.CreatedAt,
	}
}

// ToMerchantDiscountFromGetRow converts GetMerchantDiscountRow to commerce/merchant domain type.
func ToMerchantDiscountFromGetRow(row merchantstore.GetMerchantDiscountRow) merchant.MerchantDiscount {
	return toMerchantDiscount(row.ID, row.MerchantID, row.ProductID, row.CategoryID, row.Type, row.Value, row.Description, row.ValidFrom, row.ValidTo, row.CreatedAt)
}

// ToMerchantDiscountFromListRow converts ListDiscountsByMerchantRow to commerce/merchant domain type.
func ToMerchantDiscountFromListRow(row merchantstore.ListDiscountsByMerchantRow) merchant.MerchantDiscount {
	return toMerchantDiscount(row.ID, row.MerchantID, row.ProductID, row.CategoryID, row.Type, row.Value, row.Description, row.ValidFrom, row.ValidTo, row.CreatedAt)
}

func toMerchantDiscount(id, merchantID, productID, categoryID uuid.UUID, discountType merchantstore.DiscountType, value pgtype.Numeric, desc pgtype.Text, validFrom, validTo, createdAt time.Time) merchant.MerchantDiscount {
	var descPtr *string
	if desc.Valid {
		descPtr = &desc.String
	}
	val, _ := value.Float64Value()
	return merchant.MerchantDiscount{
		ID:          id,
		MerchantID:  merchantID,
		Type:        merchant.DiscountType(discountType),
		Value:       val.Float64,
		Description: descPtr,
		ValidFrom:   validFrom,
		ValidTo:     validTo,
		CreatedAt:   createdAt,
		ProductID:   productID,
		CategoryID:  categoryID,
	}
}
