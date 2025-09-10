package model

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ServiceName string     `json:"service_name" db:"service_name" example:"Yandex Plus"`
	Price       int64      `json:"price" db:"price" example:"299"`
	UserId      uuid.UUID  `json:"user_id" db:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   time.Time  `json:"start_date" db:"start_date" example:"2023-10-01T00:00:00Z"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date" example:"2024-10-01T00:00:00Z"`
}
