package database

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/nabilulilalbab/TopsisByme/models"
)

var DB *gorm.DB

func InitDB() {
	// Menggunakan variabel lingkungan standar dari Railway
	host := os.Getenv("MYSQLHOST")
	user := os.Getenv("MYSQLUSER")
	password := os.Getenv("MYSQLPASSWORD")
	dbname := os.Getenv("MYSQLDATABASE")
	port := os.Getenv("MYSQLPORT")

	// Memastikan port tidak kosong, jika aplikasi dijalankan lokal
	if port == "" {
		port = "3306" // Port default MySQL
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		password,
		host,
		port,
		dbname,
	)

	var db *gorm.DB
	var err error

	// Logika retry Anda sudah bagus, tidak perlu diubah
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, errPing := db.DB()
			if errPing == nil {
				errPing = sqlDB.Ping()
			}
			if errPing == nil {
				fmt.Println("Successfully connected to the database!")
				break // koneksi dan ping berhasil
			} else {
				err = errPing
			}
		}
		fmt.Printf("Failed to connect to database, retrying in 3s... error: %v\n", err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		panic("Failed to connect to database after retries: " + err.Error())
	}

	fmt.Println("Running database migrations...")
	db.AutoMigrate(
		&models.TopsisCalculation{},
		&models.Alternative{},
		&models.CriteriaValue{},
		&models.IdealSolution{},
		&models.User{},
	)
	fmt.Println("Migrations completed.")

	DB = db
}
