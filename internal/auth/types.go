package auth

import (
	authstore "github.com/horiondreher/go-web-api-boilerplate/internal/auth/store"
)

type DBTX = authstore.DBTX

type ActorRole = authstore.ActorRole
type AssignActorRoleParams = authstore.AssignActorRoleParams
type GetActorRoleParams = authstore.GetActorRoleParams
type TouchActorRoleAssignedAtParams = authstore.TouchActorRoleAssignedAtParams
