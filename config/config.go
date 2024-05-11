package config

import (
	"database/sql"
	"fmt"
	"os"

	godotenv "github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type Config struct {
	DB_HOST    string
	DB_PORT    string
	DB_USER    string
	DB_PASS    string
	DB_NAME    string
	JWT_KEY    string
	JWT_ISSUER string
}

var (
	SERVER_PORT string
	ENVIRONMENT string
	DB_HOST     string
	DB_PORT     string
	DB_USER     string
	DB_PASS     string
	DB_NAME     string
	JWT_KEY     string
	JWT_ISSUER  string
)

func Init() {
	godotenv.Load(".env")

	SERVER_PORT = os.Getenv("SERVER_PORT")
	ENVIRONMENT = os.Getenv("ENVIRONMENT")
	DB_PORT = os.Getenv("DB_PORT")
	DB_NAME = os.Getenv("DB_NAME")
	DB_USER = os.Getenv("DB_USER")
	DB_PASS = os.Getenv("DB_PASS")
	DB_HOST = os.Getenv("DB_HOST")
	JWT_KEY = os.Getenv("JWT_KEY")
	JWT_ISSUER = os.Getenv("DB_HOST")
}

func DbCon() (*sql.DB, error) {
	// Load konfigurasi dari file .env
	Init()

	// Format string koneksi ke PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME)

	// Buat koneksi ke database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Uji koneksi ke database
	err = db.Ping()
	if err != nil {
		db.Close() // Tutup koneksi jika ping gagal
		return nil, err
	}

	fmt.Println("Berhasil terhubung ke database PostgreSQL")

	return db, nil
}
