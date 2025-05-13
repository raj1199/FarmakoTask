package api

import (
	"net/http"
	"time"

	"coupon-system/internal/models"
	"coupon-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	couponService *service.CouponService
}

func NewHandler(couponService *service.CouponService) *Handler {
	return &Handler{
		couponService: couponService,
	}
}

// @Summary Create a new coupon
// @Description Create a new coupon with the given parameters
// @Tags coupons
// @Accept json
// @Produce json
// @Param coupon body CreateCouponRequest true "Coupon creation request"
// @Success 201 {object} models.Coupon
// @Failure 400 {object} ErrorResponse
// @Router /admin/coupons [post]
func (h *Handler) CreateCoupon(c *gin.Context) {
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := service.CreateCouponInput{
		Code:                 req.Code,
		ExpiryDate:           req.ExpiryDate,
		UsageType:            models.UsageType(req.UsageType),
		DiscountType:         models.DiscountType(req.DiscountType),
		DiscountValue:        req.DiscountValue,
		MinOrderValue:        req.MinOrderValue,
		MaxUsagePerUser:      req.MaxUsagePerUser,
		ValidTimeWindow:      req.ValidTimeWindow,
		TermsAndConditions:   req.TermsAndConditions,
		ApplicableMedicines:  req.ApplicableMedicines,
		ApplicableCategories: req.ApplicableCategories,
	}

	coupon, err := h.couponService.CreateCoupon(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, coupon)
}

// @Summary Get applicable coupons
// @Description Get all applicable coupons for the given cart items
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body GetApplicableCouponsRequest true "Get applicable coupons request"
// @Success 200 {array} models.Coupon
// @Failure 400 {object} ErrorResponse
// @Router /coupons/applicable [get]
func (h *Handler) GetApplicableCoupons(c *gin.Context) {
	var req GetApplicableCouponsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	coupons, err := h.couponService.GetApplicableCoupons(
		c.Request.Context(),
		req.CartItems,
		req.OrderTotal,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applicable_coupons": coupons,
	})
}

// @Summary Validate a coupon
// @Description Validate a coupon for the given cart items
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body ValidateCouponRequest true "Validate coupon request"
// @Success 200 {object} service.ValidateCouponOutput
// @Failure 400 {object} ErrorResponse
// @Router /coupons/validate [post]
func (h *Handler) ValidateCoupon(c *gin.Context) {
	var req ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	input := service.ValidateCouponInput{
		Code:       req.CouponCode,
		CartItems:  req.CartItems,
		OrderTotal: req.OrderTotal,
		UserID:     userID.(uuid.UUID),
		Timestamp:  time.Now(),
	}

	result, err := h.couponService.ValidateCoupon(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type CreateCouponRequest struct {
	Code                 string             `json:"code" binding:"required"`
	ExpiryDate           time.Time          `json:"expiry_date" binding:"required"`
	UsageType            string             `json:"usage_type" binding:"required,oneof=one_time multi_use time_based"`
	DiscountType         string             `json:"discount_type" binding:"required,oneof=percentage fixed"`
	DiscountValue        float64            `json:"discount_value" binding:"required,gt=0"`
	MinOrderValue        float64            `json:"min_order_value" binding:"gte=0"`
	MaxUsagePerUser      int                `json:"max_usage_per_user" binding:"required,gte=1"`
	ValidTimeWindow      *models.TimeWindow `json:"valid_time_window"`
	TermsAndConditions   string             `json:"terms_and_conditions"`
	ApplicableMedicines  []models.Medicine  `json:"applicable_medicines"`
	ApplicableCategories []models.Category  `json:"applicable_categories"`
}

type GetApplicableCouponsRequest struct {
	CartItems  []models.Medicine `json:"cart_items" binding:"required"`
	OrderTotal float64           `json:"order_total" binding:"required,gte=0"`
}

type ValidateCouponRequest struct {
	CouponCode string            `json:"coupon_code" binding:"required"`
	CartItems  []models.Medicine `json:"cart_items" binding:"required"`
	OrderTotal float64           `json:"order_total" binding:"required,gte=0"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
