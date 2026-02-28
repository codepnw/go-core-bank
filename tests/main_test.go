package tests

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Test Load Env not found: %v\n", err)
	}

	var err error

	testDB, err = sql.Open("postgres", os.Getenv("DB_TEST_URL"))
	if err != nil {
		log.Fatalf("Connect Test DB failed: %v", err)
	}

	if err := testDB.Ping(); err != nil {
		log.Fatalf("Ping Test DB failed: %v", err)
	}
	log.Println("Content Test DB (corebank_test) Successfully")

	code := m.Run()

	testDB.Close()
	log.Println("Close Test DB Successfully")

	os.Exit(code)
}
