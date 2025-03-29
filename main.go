package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

func init() {
	LoadEnv()
}

func LoadEnv() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("error loading .env file: %v", err)
		}
	}
}

type ServerConfig struct {
	AllowedOrigins string `env:"ALLOWED_ORIGINS,required"`
	Port           string `env:"SERVER_PORT"`
	Timezone       string `env:"SERVER_TIMEZONE,default=America/New_York"`
	Environment    string `env:"SERVER_ENVIRONMENT,default=development"`
	AppEnv         string `env:"APP_ENV,required"`
}

type AppConfig struct {
	Server *ServerConfig
}

func GetAppConfig(ctx context.Context) *AppConfig {
	var cfg AppConfig
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatalf("error processing env vars: %v", err)
	}

	if cfg.Server.Port == "" {
		cfg.Server.Port = os.Getenv("PORT")
		cfg.Server.AppEnv = os.Getenv("APP_ENV")
		cfg.Server.AllowedOrigins = os.Getenv("ALLOWED_ORIGINS")
	}

	return &cfg
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := GetAppConfig(ctx)

	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"0.0.0.0"},
		ProxyHeader:             fiber.HeaderXForwardedFor,
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
	})

	app.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"data": "ok"})
	})

	// rand.Seed(time.Now().UnixNano())

	randomNumber := rand.Intn(10)

	app.Get("/app", func(ctx *fiber.Ctx) error {
		return ctx.JSON(
			fiber.Map{
				"server": fmt.Sprintf("Server-%d", randomNumber),
				"data":   cfg.Server.AppEnv,
			},
		)
	})

	app.Get("/new-api", func(ctx *fiber.Ctx) error {
		return ctx.JSON(
			fiber.Map{
				"server":   fmt.Sprintf("Server-%d", randomNumber),
				"boss-man": cfg.Server.AppEnv,
			},
		)
	})

	PORT := flag.String("PORT", cfg.Server.Port, "server port")
	flag.Parse()

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", *PORT)); err != nil {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	cancel()

	fmt.Println("Gracefully shutting down...")
	app.Shutdown()

	fmt.Println("Running cleanup tasks...")

	fmt.Println("Fiber was successful shutdown.")
}
