package models

type Model interface {
	getSchema() string
}

type BaseModel struct {
	CreatedAt
}
