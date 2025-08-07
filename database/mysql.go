package database

import (
	"database/sql"
	"fmt"
	"log"

	"webhook-listener-mekarisign/config"
	"webhook-listener-mekarisign/logger"

	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectMySQL(cfg *config.Config) {
	// Ambil konfigurasi dari environment variable atau file config
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FJakarta",
		cfg.DatabaseMysqlUser,
		cfg.DatabaseMysqlPassword,
		cfg.DatabaseMysqlHost,
		cfg.DatabaseMysqlPort,
		cfg.DatabaseMysqlDatabase,
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.NewLogger().Fatal("Gagal buka koneksi ke MySQL:", zap.Error(err))
	}

	// Coba ping database
	if err = DB.Ping(); err != nil {
		logger.NewLogger().Fatal("Gagal ping ke MySQL:", zap.Error(err))
	}

	log.Println("Connected to MySQL")
}
