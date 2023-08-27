package database

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"log"
	"sync"

	"goqrs/envs"

	"gorm.io/gorm"
)

type key int

var (
	context_connextion_key key
	connection             *gorm.DB
	once                   sync.Once
)

func logMode() logger.LogLevel {
	value := envs.FindEnv("GOQRS_DB_LOGS", "silent")
	switch value {
	case "info":
		return logger.Info
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	}
	return logger.Silent
}
func PrepareConnection() (err error) {
	once.Do(func() {
		host := envs.FindEnv("GOQRS_DB_HOST", "localhost")
		port := envs.FindEnv("GOQRS_DB_PORT", "5432")
		user := envs.FindEnv("GOQRS_DB_USER", "postgres")
		password := envs.FindEnv("GOQRS_DB_PASSWORD", "root")
		dbname := envs.FindEnv("GOQRS_DB_NAME", "tickets_system_db")

		const layer = "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
		dsn := fmt.Sprintf(layer, host, user, password, dbname, port)

		connection, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			SkipDefaultTransaction: true,
			Logger:                 logger.Default.LogMode(logMode()),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		})
	})
	return err
}

func Conn(ctx context.Context) *gorm.DB {
	value := ctx.Value(context_connextion_key)
	if value == nil {
		return connection.WithContext(ctx)
	}
	connection, ok := value.(*gorm.DB)
	if !ok {
		log.Println("WARN:", "connection invalid type:", reflect.TypeOf(value))
		return connection.WithContext(ctx)
	}
	return connection
}

func WithTx(ctx context.Context, fc func(ctx context.Context) error) error {
	if ctx.Value(context_connextion_key) != nil {
		return fc(ctx) // returns the current connection
	}
	return Conn(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(context.WithValue(ctx, context_connextion_key, tx))
	})
}
