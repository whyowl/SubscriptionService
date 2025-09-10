package model

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ServiceName string     `json:"service_name" db:"service_name"`
	Price       int64      `json:"price" db:"price"`
	UserId      uuid.UUID  `json:"user_id" db:"user_id"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
}
