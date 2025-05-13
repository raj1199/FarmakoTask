package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UsageType string
type DiscountType string

const (
	OneTime   UsageType = "one_time"
	MultiUse  UsageType = "multi_use"
	TimeBased UsageType = "time_based"

	PercentageDiscount DiscountType = "percentage"
	FixedDiscount      DiscountType = "fixed"
)

type Coupon struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Code               string         `gorm:"uniqueIndex;not null" json:"code" validate:"required"`
	ExpiryDate         time.Time      `gorm:"not null" json:"expiry_date" validate:"required,gt=now"`
	UsageType          UsageType      `gorm:"not null" json:"usage_type" validate:"required,oneof=one_time multi_use time_based"`
	DiscountType       DiscountType   `gorm:"not null" json:"discount_type" validate:"required,oneof=percentage fixed"`
	DiscountValue      float64        `gorm:"not null" json:"discount_value" validate:"required,gt=0"`
	MinOrderValue      float64        `gorm:"not null" json:"min_order_value" validate:"gte=0"`
	MaxUsagePerUser    int            `gorm:"not null" json:"max_usage_per_user" validate:"required,gte=1"`
	ValidTimeWindow    *TimeWindow    `gorm:"embedded" json:"valid_time_window,omitempty"`
	TermsAndConditions string         `gorm:"type:text" json:"terms_and_conditions"`
	IsActive           bool           `gorm:"default:true" json:"is_active"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	ApplicableMedicines  []Medicine    `gorm:"many2many:coupon_medicines;" json:"applicable_medicines,omitempty"`
	ApplicableCategories []Category    `gorm:"many2many:coupon_categories;" json:"applicable_categories,omitempty"`
	Usages               []CouponUsage `gorm:"foreignKey:CouponID" json:"-"`
}

type TimeWindow struct {
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

type Medicine struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Price    float64   `json:"price"`
}

type Category struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name string    `json:"name"`
}

type CouponUsage struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	CouponID  uuid.UUID `gorm:"type:uuid;not null" json:"coupon_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	UsedAt    time.Time `gorm:"not null" json:"used_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Coupon) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (c *Coupon) IsValid(orderTotal float64, currentTime time.Time) bool {
	if !c.IsActive {
		return false
	}

	if currentTime.After(c.ExpiryDate) {
		return false
	}

	if orderTotal < c.MinOrderValue {
		return false
	}

	if c.ValidTimeWindow != nil {
		if currentTime.Before(c.ValidTimeWindow.StartTime) || currentTime.After(c.ValidTimeWindow.EndTime) {
			return false
		}
	}

	return true
}

func (c *Coupon) CalculateDiscount(orderTotal float64) float64 {
	if c.DiscountType == PercentageDiscount {
		return orderTotal * (c.DiscountValue / 100)
	}
	return c.DiscountValue
}
