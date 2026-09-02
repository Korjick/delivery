package main

import (
	"context"
	"delivery/cmd"
	"delivery/internal/generated/servers"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	config := getConfigs()

	compositionRoot := cmd.NewCompositionRoot(config)
	defer compositionRoot.CloseAll()

	db, err := compositionRoot.OpenPostgres()
	if err != nil {
		log.Fatalf("open PostgreSQL: %v", err)
	}
	handler, err := compositionRoot.NewHTTPHandler(db)
	if err != nil {
		log.Fatalf("create HTTP handler: %v", err)
	}
	_, err = compositionRoot.NewDeliveryJobs(db)
	if err != nil {
		log.Fatalf("create delivery jobs: %v", err)
	}
	basketConsumer, err := compositionRoot.NewBasketConfirmedConsumer(db)
	if err != nil {
		log.Fatalf("create Basket confirmed consumer: %v", err)
	}
	go func() {
		if err := basketConsumer.Consume(); err != nil {
			log.Printf("Basket confirmed consumer stopped: %v", err)
		}
	}()

	startWebServer(handler, config.HttpPort, config.HttpCorsOrigins)
}

func getConfigs() cmd.Config {
	// Attempt to load .env file if present, but do not fail if absent (e.g. in containerized env)
	_ = godotenv.Load(".env")

	return cmd.Config{
		HttpPort:               getEnv("HTTP_PORT", "8082"),
		DbHost:                 getEnv("DB_HOST", "localhost"),
		DbPort:                 getEnv("DB_PORT", "5432"),
		DbUser:                 getEnv("DB_USER", "username"),
		DbPassword:             getEnv("DB_PASSWORD", "secret"),
		DbName:                 getEnv("DB_NAME", "delivery"),
		DbSslMode:              getEnv("DB_SSLMODE", "disable"),
		GeoServiceGrpcHost:     getEnv("GEO_SERVICE_GRPC_HOST", "localhost:5004"),
		KafkaHost:              getEnv("KAFKA_HOST", "localhost:9092"),
		KafkaConsumerGroup:     getEnv("KAFKA_CONSUMER_GROUP", "delivery-service-group"),
		KafkaBasketEventsTopic: getEnv("KAFKA_BASKET_EVENTS_TOPIC", "basket.confirmed"),
		KafkaOrderEventsTopic:  getEnv("KAFKA_ORDER_EVENTS_TOPIC", "order.status.changed"),
		HttpCorsOrigins:        getEnv("HTTP_CORS_ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func startWebServer(handler servers.ServerInterface, port, corsOrigins string) {
	e := echo.New()

	// Configure CORS middleware
	var origins []string
	for _, o := range strings.Split(corsOrigins, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})

	servers.RegisterHandlers(e, handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(fmt.Sprintf("0.0.0.0:%s", port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Print("Shutting down service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
