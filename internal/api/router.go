package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "alert-system/docs/swagger"
)

// NewRouter wires the HTTP routes and Swagger UI.
func NewRouter(database *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	handler := &Handler{db: database}

	r.Get("/healthz", handler.Health)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(router chi.Router) {
		router.Post("/devices", handler.CreateDevice)
		router.Get("/devices", handler.ListDevices)
		router.Get("/devices/{id}", handler.GetDevice)
		router.Post("/devices/{id}/heartbeat", handler.DeviceHeartbeat)
		router.Post("/devices/synchronize", handler.SynchronizeDeviceStatuses)

		router.Get("/alarms", handler.ListAlarms)
		router.Post("/alarms", handler.CreateAlarm)
		router.Post("/alarms/{id}/acknowledge", handler.AcknowledgeAlarm)
		router.Post("/alarms/{id}/resolve", handler.ResolveAlarm)
	})

	return r
}
