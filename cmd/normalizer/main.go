package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/muasiq/universal-inverter-data-normalizer/internal/api"
	"github.com/muasiq/universal-inverter-data-normalizer/internal/config"
	"github.com/muasiq/universal-inverter-data-normalizer/internal/normalizer"
	"github.com/muasiq/universal-inverter-data-normalizer/internal/provider"

	// Register all providers (side-effect imports)
	_ "github.com/muasiq/universal-inverter-data-normalizer/internal/provider/huawei"
	_ "github.com/muasiq/universal-inverter-data-normalizer/internal/provider/saj"
	_ "github.com/muasiq/universal-inverter-data-normalizer/internal/provider/sma"
	_ "github.com/muasiq/universal-inverter-data-normalizer/internal/provider/sungrow"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup logging
	setupLogging(cfg.Logging.Level, cfg.Logging.Format)

	log.Info().
		Str("version", "1.0.0").
		Int("providers_configured", len(cfg.Providers)).
		Msg("Starting Universal Inverter Data Normalizer")

	// List available provider types
	available := provider.DefaultRegistry.ListProviders()
	log.Info().Strs("available_providers", available).Msg("Registered provider types")

	// Create normalizer engine
	engine := normalizer.NewEngine()
	defer engine.Close()

	// Initialize and register providers
	ctx := context.Background()
	for _, pc := range cfg.Providers {
		if !pc.Enabled {
			log.Info().Str("provider", pc.Name).Msg("Provider disabled, skipping")
			continue
		}

		p, err := provider.DefaultRegistry.Create(pc.Type)
		if err != nil {
			log.Error().Err(err).Str("type", pc.Type).Str("name", pc.Name).Msg("Failed to create provider")
			continue
		}

		if err := p.Initialize(ctx, pc); err != nil {
			log.Error().Err(err).Str("provider", pc.Name).Msg("Failed to initialize provider")
			continue
		}

		engine.RegisterProvider(p)
		log.Info().Str("provider", p.Name()).Msg("Provider registered successfully")
	}

	// Start API server
	srv := api.NewServer(engine, cfg.Server.Addr())

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")
		engine.Close()
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

func setupLogging(level, format string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
