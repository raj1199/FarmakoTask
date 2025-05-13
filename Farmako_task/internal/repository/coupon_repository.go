package repository

import (
	"context"
	"errors"
	"time"

	"coupon-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) Create(ctx context.Context, coupon *models.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

func (r *CouponRepository) GetByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := r.db.WithContext(ctx).
		Preload("ApplicableMedicines").
		Preload("ApplicableCategories").
		Where("code = ? AND is_active = true", code).
		First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *CouponRepository) GetApplicableCoupons(ctx context.Context, cartItems []models.Medicine, orderTotal float64) ([]models.Coupon, error) {
	var coupons []models.Coupon
	now := time.Now()

	// Get all active coupons that haven't expired and meet the minimum order value
	query := r.db.WithContext(ctx).
		Preload("ApplicableMedicines").
		Preload("ApplicableCategories").
		Where("is_active = true AND expiry_date > ? AND min_order_value <= ?", now, orderTotal)

	err := query.Find(&coupons).Error
	if err != nil {
		return nil, err
	}

	// Filter coupons based on medicine and category restrictions
	var applicableCoupons []models.Coupon
	for _, coupon := range coupons {
		if isApplicableToCoupon(coupon, cartItems) {
			applicableCoupons = append(applicableCoupons, coupon)
		}
	}

	return applicableCoupons, nil
}

func (r *CouponRepository) GetUserCouponUsage(ctx context.Context, couponID, userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.CouponUsage{}).
		Where("coupon_id = ? AND user_id = ?", couponID, userID).
		Count(&count).Error
	return int(count), err
}

func (r *CouponRepository) RecordCouponUsage(ctx context.Context, usage *models.CouponUsage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if the coupon is still valid
		var coupon models.Coupon
		if err := tx.WithContext(ctx).Where("id = ? AND is_active = true", usage.CouponID).First(&coupon).Error; err != nil {
			return err
		}

		// For one-time use coupons, check if it's been used before
		if coupon.UsageType == models.OneTime {
			var count int64
			if err := tx.WithContext(ctx).Model(&models.CouponUsage{}).
				Where("coupon_id = ? AND user_id = ?", usage.CouponID, usage.UserID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New("one-time coupon already used")
			}
		}

		// For multi-use coupons, check usage limit
		if coupon.UsageType == models.MultiUse {
			var count int64
			if err := tx.WithContext(ctx).Model(&models.CouponUsage{}).
				Where("coupon_id = ? AND user_id = ?", usage.CouponID, usage.UserID).
				Count(&count).Error; err != nil {
				return err
			}
			if int(count) >= coupon.MaxUsagePerUser {
				return errors.New("coupon usage limit exceeded")
			}
		}

		// Record the usage
		return tx.WithContext(ctx).Create(usage).Error
	})
}

func isApplicableToCoupon(coupon models.Coupon, cartItems []models.Medicine) bool {
	if len(coupon.ApplicableMedicines) == 0 && len(coupon.ApplicableCategories) == 0 {
		return true
	}

	// Check if any cart item matches the coupon's medicine restrictions
	for _, item := range cartItems {
		// Check direct medicine match
		for _, medicine := range coupon.ApplicableMedicines {
			if item.ID == medicine.ID {
				return true
			}
		}

		// Check category match
		for _, category := range coupon.ApplicableCategories {
			if item.Category == category.Name {
				return true
			}
		}
	}

	return false
}
