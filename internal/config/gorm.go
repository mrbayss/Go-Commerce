package config

import (
	"fmt"
	"time"

	"github.com/mrbayss/Go-Commerce/internal/entity"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
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
func NewDatabase(viper *viper.Viper, log *logrus.Logger) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("postgres.host"),
		viper.GetString("postgres.port"),
		viper.GetString("postgres.user"),
		viper.GetString("postgres.pass"),
		viper.GetString("postgres.db"),
		viper.GetString("postgres.sslmode"),
	)

	// Configure GORM logger based on environment
	var gormLogger gormlogger.Interface
	env := viper.GetString("app.environment")
	if env == "development" {
		gormLogger = gormlogger.Default.LogMode(gormlogger.Info)
	} else {
		gormLogger = gormlogger.Discard
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

	// Connection pool defaults
	maxOpenConns := viper.GetInt("postgres.max_open_conns")
	if maxOpenConns <= 0 {
		maxOpenConns = 25
	}
	maxIdleConns := viper.GetInt("postgres.max_idle_conns")
	if maxIdleConns <= 0 {
		maxIdleConns = 10
	}
	connMaxLifetime := viper.GetDuration("postgres.conn_max_lifetime")
	if connMaxLifetime <= 0 {
		connMaxLifetime = 5 * time.Minute
	}
	connMaxIdleTime := viper.GetDuration("postgres.conn_max_idle_time")
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = 1 * time.Minute
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// Auto migration check
	if viper.GetBool("postgres.auto_migrate") {
		if err := AutoMigrate(db); err != nil {
			log.WithError(err).Fatal("Failed to auto-migrate database")
		}
	}

	return db
}
