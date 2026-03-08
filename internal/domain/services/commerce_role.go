package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/jackc/pgx/v5"
)

func (service *CommerceManager) hasRole(ctx context.Context, actorID uuid.UUID, roleType pgsqlc.RoleType, merchantID uuid.UUID) (bool, error) {
	var one int
	var err error
	if merchantID == uuid.Nil {
		err = service.db.QueryRow(ctx, `
            SELECT 1
            FROM actors a
            JOIN actor_roles ar ON ar.actor_id = a.id AND ar.merchant_id = a.merchant_id
            JOIN roles r ON r.id = ar.role_id AND r.merchant_id = a.merchant_id
            WHERE a.id = $1 AND r.role_type = $2
            LIMIT 1
        `, actorID, roleType).Scan(&one)
	} else {
		err = service.db.QueryRow(ctx, `
            SELECT 1
            FROM actors a
            JOIN actor_roles ar ON ar.actor_id = a.id AND ar.merchant_id = a.merchant_id
            JOIN roles r ON r.id = ar.role_id AND r.merchant_id = a.merchant_id
            WHERE a.id = $1 AND a.merchant_id = $2 AND r.role_type = $3
            LIMIT 1
        `, actorID, merchantID, roleType).Scan(&one)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (service *CommerceManager) getRoleIDByType(ctx context.Context, merchantID uuid.UUID, roleType pgsqlc.RoleType) (uuid.UUID, error) {
	var roleID uuid.UUID
	err := service.db.QueryRow(ctx, `
        SELECT id
        FROM roles
        WHERE merchant_id = $1 AND role_type = $2
        LIMIT 1
    `, merchantID, roleType).Scan(&roleID)
	if err != nil {
		return uuid.Nil, err
	}
	return roleID, nil
}
