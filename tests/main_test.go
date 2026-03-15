package tests

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	testDB    *sql.DB
	testRedis *redis.Client
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Test Load Env not found: %v\n", err)
	}
	
	// Init DB
	db, err := sql.Open("postgres", os.Getenv("DB_TEST_URL"))
	if err != nil {
		log.Fatalf("Connect Test DB failed: %v", err)
	}
	testDB = db
	
	defer db.Close()

	if err := testDB.PingContext(ctx); err != nil {
		log.Fatalf("Ping Test DB failed: %v", err)
	}
	log.Println("Connect Test DB (corebank_test) Successfully")

	// Init Redis
	options := &redis.Options{
		Addr:     os.Getenv("REDIS_TEST_ADDR"),
		Password: os.Getenv("REDIS_TEST_PASSWORD"),
		DB:       0, // Default DB
	}
	rdb := redis.NewClient(options)
	testRedis = rdb

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Ping Redis failed: %v", err)
	}

	os.Exit(m.Run())
}
