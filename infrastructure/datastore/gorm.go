package datastore

import (
	"fmt"

	"github.com/mrbayss/Go-Commerce/infrastructure/config"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AutoMigrate migrates all entity schemas into the database.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.Address{},
		&entity.Store{},
		&entity.Category{},
		&entity.Product{},
		&entity.ProductVariant{},
		&entity.ProductImage{},
		&entity.Cart{},
		&entity.CartItem{},
		&entity.Voucher{},
		&entity.Order{},
		&entity.StoreOrder{},
		&entity.OrderItem{},
		&entity.Payment{},
		&entity.Shipment{},
		&entity.Wallet{},
		&entity.WalletTransaction{},
		&entity.VoucherUsage{},
		&entity.Campaign{},
		&entity.FlashSaleItem{},
		&entity.Review{},
		&entity.Conversation{},
		&entity.ConversationParticipant{},
		&entity.Message{},
		&entity.Notification{},
	)
}

// NewDatabase creates a new GORM database connection and configures the connection pool.
func NewDatabase(cfg *config.Config, log *logrus.Logger) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Pass,
		cfg.Postgres.DB,
		cfg.Postgres.SSLMode,
	)

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if cfg.App.Environment == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Discard
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         gormLogger,
	})
	if err != nil {
		panic(fmt.Errorf("failed to connect database: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Errorf("failed to get sql.DB: %v", err))
	}

	// Connection pool defaults (from Config.Load())
	sqlDB.SetMaxOpenConns(cfg.Postgres.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Postgres.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Postgres.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Postgres.ConnMaxIdleTime)

	// Auto migration check
	if cfg.Postgres.AutoMigrate {
		if err := AutoMigrate(db); err != nil {
			log.WithError(err).Fatal("Failed to auto-migrate database")
		}
	}

	return db
}
