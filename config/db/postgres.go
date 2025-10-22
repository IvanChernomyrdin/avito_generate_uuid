package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var DB *sql.DB

func NewConnectionPostgres(DSN string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("pgx", DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настраиваем пул соединений
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Проверяем соединение
	if err = DB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return DB, nil
}

func CloseConnection(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func MigrationsDatabase(DB *sql.DB) error {
	if _, err := os.Stat("migrations/postgres"); os.IsNotExist(err) {
		return nil // Папки нет - это нормально
	}

	// запускаем миграции
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed create driver migrations: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed create migrations: %w", err)
	}
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			// Нет новых миграций - это не ошибка
			return nil
		}
		return fmt.Errorf("failed confirm migrations: %w", err)
	}
	return nil
}

func PingDatabase() error {
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("The database isn't responding")
	}
	return nil
}
