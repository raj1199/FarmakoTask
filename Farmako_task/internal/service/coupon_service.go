package service

import (
	"context"
	"time"

	"coupon-system/internal/models"
	"coupon-system/internal/repository"

	"github.com/google/uuid"
)

type CouponService struct {
	repo *repository.CouponRepository
}

func NewCouponService(repo *repository.CouponRepository) *CouponService {
	return &CouponService{repo: repo}
}

type CreateCouponInput struct {
	Code                 string
	ExpiryDate           time.Time
	UsageType            models.UsageType
	DiscountType         models.DiscountType
	DiscountValue        float64
	MinOrderValue        float64
	MaxUsagePerUser      int
	ValidTimeWindow      *models.TimeWindow
	TermsAndConditions   string
	ApplicableMedicines  []models.Medicine
	ApplicableCategories []models.Category
}

func (s *CouponService) CreateCoupon(ctx context.Context, input CreateCouponInput) (*models.Coupon, error) {
	coupon := &models.Coupon{
		ID:                   uuid.New(),
		Code:                 input.Code,
		ExpiryDate:           input.ExpiryDate,
		UsageType:            input.UsageType,
		DiscountType:         input.DiscountType,
		DiscountValue:        input.DiscountValue,
		MinOrderValue:        input.MinOrderValue,
		MaxUsagePerUser:      input.MaxUsagePerUser,
		ValidTimeWindow:      input.ValidTimeWindow,
		TermsAndConditions:   input.TermsAndConditions,
		ApplicableMedicines:  input.ApplicableMedicines,
		ApplicableCategories: input.ApplicableCategories,
		IsActive:             true,
	}

	if err := s.repo.Create(ctx, coupon); err != nil {
		return nil, err
	}

	return coupon, nil
}

type ValidateCouponInput struct {
	Code       string
	CartItems  []models.Medicine
	OrderTotal float64
	UserID     uuid.UUID
	Timestamp  time.Time
}

type ValidateCouponOutput struct {
	IsValid         bool
	ItemsDiscount   float64
	ChargesDiscount float64
	Message         string
}

func (s *CouponService) ValidateCoupon(ctx context.Context, input ValidateCouponInput) (*ValidateCouponOutput, error) {
	coupon, err := s.repo.GetByCode(ctx, input.Code)
	if err != nil {
		return nil, err
	}

	if coupon == nil {
		return &ValidateCouponOutput{
			IsValid: false,
			Message: "coupon not found",
		}, nil
	}

	// Basic validation
	if !coupon.IsValid(input.OrderTotal, input.Timestamp) {
		return &ValidateCouponOutput{
			IsValid: false,
			Message: "coupon is not valid for this order",
		}, nil
	}

	// Check if the coupon is applicable to the cart items
	if !isApplicableToCoupon(*coupon, input.CartItems) {
		return &ValidateCouponOutput{
			IsValid: false,
			Message: "coupon is not applicable to any items in cart",
		}, nil
	}

	// Check usage limits
	usageCount, err := s.repo.GetUserCouponUsage(ctx, coupon.ID, input.UserID)
	if err != nil {
		return nil, err
	}

	if coupon.UsageType == models.OneTime && usageCount > 0 {
		return &ValidateCouponOutput{
			IsValid: false,
			Message: "one-time coupon already used",
		}, nil
	}

	if coupon.UsageType == models.MultiUse && usageCount >= coupon.MaxUsagePerUser {
		return &ValidateCouponOutput{
			IsValid: false,
			Message: "coupon usage limit exceeded",
		}, nil
	}

	// Calculate discount
	discount := coupon.CalculateDiscount(input.OrderTotal)

	return &ValidateCouponOutput{
		IsValid:         true,
		ItemsDiscount:   discount,
		ChargesDiscount: 0, // Can be extended for delivery fee discounts
		Message:         "coupon applied successfully",
	}, nil
}

func (s *CouponService) GetApplicableCoupons(ctx context.Context, cartItems []models.Medicine, orderTotal float64) ([]models.Coupon, error) {
	return s.repo.GetApplicableCoupons(ctx, cartItems, orderTotal)
}

func (s *CouponService) RecordCouponUsage(ctx context.Context, couponID, userID, orderID uuid.UUID) error {
	usage := &models.CouponUsage{
		ID:        uuid.New(),
		CouponID:  couponID,
		UserID:    userID,
		OrderID:   orderID,
		UsedAt:    time.Now(),
		CreatedAt: time.Now(),
	}

	return s.repo.RecordCouponUsage(ctx, usage)
}

// Helper function to check if a coupon is applicable to cart items
func isApplicableToCoupon(coupon models.Coupon, cartItems []models.Medicine) bool {
	if len(coupon.ApplicableMedicines) == 0 && len(coupon.ApplicableCategories) == 0 {
		return true
	}

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
