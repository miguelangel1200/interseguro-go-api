package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	infrahttp "interseguro/go-api/internal/infrastructure/http"
	"interseguro/go-api/internal/infrastructure/remote"
	"interseguro/go-api/internal/application/services"
)

func main() {
	// Puerto en el que escucha la API. Configurable vía variable de entorno.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// URL de la API Node.js que recibe el resultado. Configurable vía entorno.
	nodeAPIURL := os.Getenv("NODE_API_URL")
	if nodeAPIURL == "" {
		nodeAPIURL = "http://localhost:3000"
	}

	// Configuración JWT compartida con la API Node.js (mismo secreto/firma HS256).
	authCfg := infrahttp.AuthConfig{
		Secret:   getenv("JWT_SECRET", "interseguro-dev-secret"),
		Issuer:   getenv("JWT_ISSUER", "interseguro"),
		Audience: getenv("JWT_AUDIENCE", "interseguro-api"),
	}

	// Composición de dependencias (wiring) según arquitectura hexagonal:
	// adaptador de salida -> servicio (caso de uso) -> adaptador de entrada.
	nodeClient := remote.NewNodeClient(nodeAPIURL)                       // puerto de salida
	matrixService := services.NewMatrixService(nodeClient)               // caso de uso
	matrixHandler := infrahttp.NewMatrixHandler(matrixService)        // puerto de entrada

	app := fiber.New(fiber.Config{
		AppName:      "Interseguro Go API",
		ErrorHandler: infrahttp.ErrorHandler,
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Authorization",
	}))

	// Rutas de salud y de operación principal (protegida por JWT).
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "go-api"})
	})
	app.Post("/process", infrahttp.JWTAuth(authCfg), matrixHandler.Process)

	log.Printf("Go API listening on :%s, forwarding to Node API at %s", port, nodeAPIURL)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// getenv devuelve el valor de una variable de entorno o un valor por defecto.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
