package config

import (
	"database/sql"
	"log"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

func Dbconnect() *sql.DB {
	dsn := "root:rootwdp@tcp(127.0.0.1:3306)/employee"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("error in getting connection :", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to the database!")
	return db
}

func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := client.Ping().Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return client
}
