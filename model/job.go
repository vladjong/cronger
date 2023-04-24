package model

import "github.com/google/uuid"

type Job struct {
	Id         uuid.UUID
	Title      string
	Tag        string
	Expression string
	IsWork     bool
}
