package store

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func getDSN() string {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "hackathon"
	}
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslMode)
}

func NewDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", getDSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func generateUUID() string {
	return uuid.New().String()
}

func randRead(b []byte) (int, error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		f, err = os.Open("C:\\Windows\\System32\\drivers\\etc\\hosts")
		if err != nil {
			return 0, err
		}
	}
	defer f.Close()
	return f.Read(b)
}