package data

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	MinTide, MaxTide *float64
}

func PostgresFromEnvOrDie() *gorm.DB {
	pw := os.Getenv("PGPASSWORD")
	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	dsn := fmt.Sprintf("host=%s user=postgres password=%s dbname=surfdash port=%s sslmode=disable TimeZone=America/Los_Angeles",
		host,
		pw,
		port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	db.AutoMigrate(&User{})
	return db
}
