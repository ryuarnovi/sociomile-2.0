package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisHost  string
	RedisPort  string
	RabbitHost string
	RabbitPort string
	RabbitUser string
	RabbitPass string
	JWTSecret  string
	ServerPort string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "sociomile"),
		DBPassword: getEnv("DB_PASSWORD", "sociomile_password"),
		DBName:     getEnv("DB_NAME", "sociomile_db"),
		RedisHost:  getEnv("REDIS_HOST", "localhost"),
		RedisPort:  getEnv("REDIS_PORT", "6379"),
		RabbitHost: getEnv("RABBITMQ_HOST", "localhost"),
		RabbitPort: getEnv("RABBITMQ_PORT", "5672"),
		RabbitUser: getEnv("RABBITMQ_USER", "guest"),
		RabbitPass: getEnv("RABBITMQ_PASSWORD", "guest"),
		JWTSecret:  getEnv("JWT_SECRET", "your-secret-key"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func NewRedisClient(cfg *Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		log.Println("Connected to Redis")
	}

	return client
}

func NewRabbitMQ(cfg *Config) (*amqp.Connection, *amqp.Channel) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitUser, cfg.RabbitPass, cfg.RabbitHost, cfg.RabbitPort)

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v", err)
		return nil, nil
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Warning: Failed to open RabbitMQ channel: %v", err)
		conn.Close()
		return nil, nil
	}

	// Declare exchanges and queues
	exchanges := []string{"conversation.events", "ticket.events"}
	for _, exchange := range exchanges {
		err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
		if err != nil {
			log.Printf("Warning: Failed to declare exchange %s: %v", exchange, err)
		}
	}

	log.Println("Connected to RabbitMQ")
	return conn, ch
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
