package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"
)

type DBTX = actorstore.DBTX

type Postgres struct {
	queries *actorstore.Queries
}

func New(db DBTX) *Postgres {
	return &Postgres{queries: actorstore.New(db)}
}

func (store *Postgres) CreateActor(ctx context.Context, arg actorstore.CreateActorParams) (actorstore.CreateActorRow, error) {
	row, err := store.queries.CreateActor(ctx, arg)
	if err != nil {
		return actorstore.CreateActorRow{}, err
	}

	return row, nil
}

func (store *Postgres) GetActor(ctx context.Context, arg actorstore.GetActorParams) (actorstore.GetActorRow, error) {
	row, err := store.queries.GetActor(ctx, arg)
	if err != nil {
		return actorstore.GetActorRow{}, err
	}

	return row, nil
}

func (store *Postgres) CreateSession(ctx context.Context, arg actorstore.CreateSessionParams) (actorstore.CreateSessionRow, error) {
	row, err := store.queries.CreateSession(ctx, arg)
	if err != nil {
		return actorstore.CreateSessionRow{}, err
	}

	return row, nil
}

func (store *Postgres) GetSession(ctx context.Context, id uuid.UUID) (actorstore.GetSessionRow, error) {
	row, err := store.queries.GetSession(ctx, id)
	if err != nil {
		return actorstore.GetSessionRow{}, err
	}

	return row, nil
}

func (store *Postgres) GetActorByUID(ctx context.Context, arg actorstore.GetActorByUIDParams) (actorstore.GetActorByUIDRow, error) {
	row, err := store.queries.GetActorByUID(ctx, arg)
	if err != nil {
		return actorstore.GetActorByUIDRow{}, err
	}

	return row, nil
}

func (store *Postgres) GetActorProfileByMerchantAndEmail(ctx context.Context, arg actorstore.GetActorProfileByMerchantAndEmailParams) (actorstore.GetActorProfileByMerchantAndEmailRow, error) {
	row, err := store.queries.GetActorProfileByMerchantAndEmail(ctx, arg)
	if err != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, err
	}

	return row, nil
}

func (store *Postgres) ListActorsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]actorstore.ListActorsByMerchantRow, error) {
	actors, err := store.queries.ListActorsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]actorstore.ListActorsByMerchantRow, 0, len(actors))
	for _, actor := range actors {

		result = append(result, actor)
	}

	return result, nil
}

func (store *Postgres) ListEmployeesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]actorstore.ListEmployeesByMerchantRow, error) {
	employees, err := store.queries.ListEmployeesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	result := make([]actorstore.ListEmployeesByMerchantRow, 0, len(employees))
	for _, employee := range employees {

		result = append(result, employee)
	}

	return result, nil
}
