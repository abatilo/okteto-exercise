package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/abatilo/okteto-exercise/cmd/podlist/server"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	flagSet := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	flagSet.Bool("debug", false, "Enable debug logging")
	flagSet.Parse(os.Args[1:])

	viper.BindPFlags(flagSet)
	viper.SetEnvPrefix("PODLIST")
	viper.AutomaticEnv()

	if viper.GetBool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	s := server.NewServer(server.WithLogger(log))

	// Register signal handlers for graceful shutdown
	done := make(chan struct{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("Shutting down gracefully")
		s.Shutdown(context.TODO())
		close(done)
	}()

	log.Info().Msg("Listening on port 8080")
	if err := s.Start(); err != http.ErrServerClosed {
		log.Error().Err(err).Msg("couldn't shut down gracefully")
	}
	<-done
	log.Info().Msg("Exiting")
}
