package store

import (
	"context"

	"github.com/horiondreher/go-web-api-boilerplate/internal/auth/store/generated"
)

type Postgres struct {
	queries *authstore.Queries
}

func New(db DBTX) *Postgres {
	return &Postgres{queries: authstore.New(db)}
}

func (store *Postgres) AssignActorRole(ctx context.Context, arg AssignActorRoleParams) (ActorRole, error) {
	role, err := store.queries.AssignActorRole(ctx, authstore.AssignActorRoleParams(arg))
	return role, err
}

func (store *Postgres) GetActorRole(ctx context.Context, arg GetActorRoleParams) (ActorRole, error) {
	role, err := store.queries.GetActorRole(ctx, authstore.GetActorRoleParams(arg))
	return role, err
}

func (store *Postgres) TouchActorRoleAssignedAt(ctx context.Context, arg TouchActorRoleAssignedAtParams) (ActorRole, error) {
	role, err := store.queries.TouchActorRoleAssignedAt(ctx, authstore.TouchActorRoleAssignedAtParams(arg))
	return role, err
}
