package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/muasiq/universal-inverter-data-normalizer/internal/normalizer"
	"github.com/rs/zerolog/log"
)

// Server wraps the HTTP server and routes.
type Server struct {
	engine *normalizer.Engine
	router chi.Router
	addr   string
}

// NewServer creates a new API server.
func NewServer(engine *normalizer.Engine, addr string) *Server {
	s := &Server{
		engine: engine,
		addr:   addr,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(RequestLogger)
	r.Use(chimiddleware.Recoverer)
	r.Use(CORS)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", s.handleHealth)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Plants
		r.Get("/plants", s.handleGetPlants)
		r.Get("/plants/{plantId}", s.handleGetPlantDetails)
		r.Get("/plants/{plantId}/devices", s.handleGetPlantDevices)
		r.Get("/plants/{plantId}/energy", s.handleGetPlantEnergy)

		// Devices
		r.Get("/devices", s.handleGetDevices)
		r.Get("/devices/{deviceId}/realtime", s.handleGetRealTimeData)
		r.Get("/devices/{deviceId}/history", s.handleGetHistoricalData)
		r.Get("/devices/{deviceId}/alarms", s.handleGetDeviceAlarms)

		// Alarms
		r.Get("/alarms", s.handleGetAllAlarms)

		// Provider info
		r.Get("/providers", s.handleGetProviders)
	})

	s.router = r
}

// Start begins listening for requests.
func (s *Server) Start() error {
	log.Info().Str("addr", s.addr).Msg("Starting API server")
	return http.ListenAndServe(s.addr, s.router)
}
