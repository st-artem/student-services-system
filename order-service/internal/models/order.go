package models

import "time"

type OrderStatus string

const (
	StatusNew       OrderStatus = "NEW"
	StatusPaid      OrderStatus = "PAID"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	Amount        float64     `gorm:"not null" json:"amount"`
	Currency      string      `gorm:"not null;default:'UAH'" json:"currency"`
	Status        OrderStatus `gorm:"not null;default:'NEW'" json:"status"`
	PaymentMethod string      `gorm:"not null" json:"paymentMethod"`
	Description   string      `json:"description"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

type CreateOrderInput struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"paymentMethod" binding:"required"`
	Description   string  `json:"description"`
}

type UpdateDescInput struct {
	Description string `json:"description" binding:"required"`
}