package model

import "github.com/google/uuid"

type Job struct {
	Id         uuid.UUID
	Tag        string
	Expression string
}
