package postgre

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/lib/pq" //nolint:revive
	"github.com/rotisserie/eris"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"wb/config"
	"wb/internal/orm/models"
)

var (
	ErrInvalidDatabaseName      = errors.New("invalid database name")
	ErrInvalidDatabaseNameChars = errors.New("database name contains invalid characters")
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	baseDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password)

	sqlDB, err := sql.Open("postgres", baseDSN)
	if err != nil {
		return nil, eris.Wrapf(err, "ошибка подключения к PostgreSQL")
	}
	defer sqlDB.Close()

	if err = sqlDB.Ping(); err != nil {
		return nil, eris.Wrapf(err, "ошибка проверки подключения")
	}

	var exists bool

	err = sqlDB.QueryRow("SELECT EXISTS (SELECT FROM pg_database WHERE datname = $1)", cfg.Database.Name).Scan(&exists)
	if err != nil {
		return nil, eris.Wrapf(err, "ошибка проверки существования базы")
	}

	if !exists {
		if cfg.Database.Name == "" || len(cfg.Database.Name) > 63 {
			return nil, eris.Wrapf(ErrInvalidDatabaseName, "database name: %s", cfg.Database.Name)
		}

		for _, char := range cfg.Database.Name {
			if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '_' {
				return nil, eris.Wrapf(ErrInvalidDatabaseNameChars, "database name: %s", cfg.Database.Name)
			}
		}

		_, err = sqlDB.Exec("CREATE DATABASE " + cfg.Database.Name)
		if err != nil {
			return nil, eris.Wrapf(err, "ошибка создания базы данных")
		}

		log.Printf("База данных %s создана", cfg.Database.Name)
	} else {
		log.Printf("База данных %s уже существует", cfg.Database.Name)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, eris.Wrapf(err, "ошибка подключения к GORM")
	}

	log.Println("Начинаем умную миграцию таблиц...")

	err = gormDB.AutoMigrate(&models.Order{})
	if err != nil {
		return nil, eris.Wrapf(err, "ошибка миграции таблицы orders")
	}

	log.Println("Таблица orders проверена и обновлена")

	err = gormDB.AutoMigrate(&models.Delivery{})
	if err != nil {
		return nil, eris.Wrapf(err, "ошибка миграции таблицы deliveries")
	}

	log.Println("Таблица deliveries проверена и обновлена")

	err = gormDB.AutoMigrate(&models.Payment{})
	if err != nil {
		return nil, eris.Wrapf(err, "ошибка миграции таблицы payments")
	}

	log.Println("Таблица payments проверена и обновлена")

	err = gormDB.AutoMigrate(&models.OrderItem{})
	if err != nil {
		return nil, eris.Wrapf(err, "ошибка миграции таблицы order_items")
	}

	log.Println("Таблица order_items проверена и обновлена")

	err = createMissingIndexes(gormDB)
	if err != nil {
		log.Printf("Предупреждение: не удалось создать некоторые индексы: %v", err)
	}

	log.Println("Миграция таблиц завершена успешно")

	return gormDB, nil
}

func createMissingIndexes(db *gorm.DB) error {
	err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_order_uid ON orders(order_uid)`).Error
	if err != nil {
		return eris.Wrapf(err, "ошибка создания индекса idx_orders_order_uid")
	}

	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_track_number ON orders(track_number)`).Error
	if err != nil {
		return eris.Wrapf(err, "ошибка создания индекса idx_orders_track_number")
	}

	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id)`).Error
	if err != nil {
		return eris.Wrapf(err, "ошибка создания индекса idx_orders_customer_id")
	}

	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_deliveries_order_id ON deliveries(order_id)`).Error
	if err != nil {
		return eris.Wrapf(err, "ошибка создания индекса idx_deliveries_order_id")
	}

	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id)`).Error
	if err != nil {
		return eris.Wrapf(err, "ошибка создания индекса idx_payments_order_id")
	}

	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id)`).Error
	if err != nil {
		return eris.Wrapf(err, "ошибка создания индекса idx_order_items_order_id")
	}

	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_order_items_chrt_id ON order_items(chrt_id)`).Error
	if err != nil {
		return eris.Wrapf(err, "ошибка создания индекса idx_order_items_chrt_id")
	}

	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_order_items_nm_id ON order_items(nm_id)`).Error
	if err != nil {
		return eris.Wrapf(err, "ошибка создания индекса idx_order_items_nm_id")
	}

	log.Println("Индексы проверены и созданы")

	return nil
}
