package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coupon-system/internal/api"
	"coupon-system/internal/models"
	"coupon-system/internal/repository"
	"coupon-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Redis
	redisClient := initRedis()

	// Initialize repositories
	couponRepo := repository.NewCouponRepository(db)

	// Initialize services
	couponService := service.NewCouponService(couponRepo)

	// Initialize handlers
	handler := api.NewHandler(couponService)

	// Initialize router
	router := setupRouter(handler)

	// Create server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func initDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=coupon_system port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schema
	err = db.AutoMigrate(
		&models.Coupon{},
		&models.Medicine{},
		&models.Category{},
		&models.CouponUsage{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initRedis() *redis.Client {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
}

func setupRouter(handler *api.Handler) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Routes
	admin := router.Group("/admin")
	{
		admin.POST("/coupons", handler.CreateCoupon)
	}

	coupons := router.Group("/coupons")
	{
		coupons.GET("/applicable", handler.GetApplicableCoupons)
		coupons.POST("/validate", handler.ValidateCoupon)
	}

	return router
}
