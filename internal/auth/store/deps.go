package store

import (
	"context"

	"github.com/horiondreher/go-web-api-boilerplate/internal/auth/store/generated"
)

type DBTX = authstore.DBTX
type ActorRole = authstore.ActorRole
type AssignActorRoleParams = authstore.AssignActorRoleParams
type GetActorRoleParams = authstore.GetActorRoleParams
type TouchActorRoleAssignedAtParams = authstore.TouchActorRoleAssignedAtParams

type _ = context.Context
