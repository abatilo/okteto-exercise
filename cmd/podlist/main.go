package main

import (
	"os"

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

	log.Info().Msg("Starting podlist")
}
