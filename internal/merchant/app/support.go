package app

import (
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	"github.com/jackc/pgx/v5/pgtype"
)

func textValue(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{String: value, Valid: true}
}

func branchFromListRow(row merchantstore.ListBranchesByMerchantRow) merchantstore.Branch {
	return merchantstore.Branch{
		ID:                 row.ID,
		MerchantID:         row.MerchantID,
		Name:               row.Name,
		Address:            row.Address,
		ContactNumber:      row.ContactNumber,
		City:               row.City,
		OpeningTimeMinutes: row.OpeningTimeMinutes,
		ClosingTimeMinutes: row.ClosingTimeMinutes,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}
