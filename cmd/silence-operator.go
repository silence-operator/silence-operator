package main

import (
	"github.com/rs/zerolog/log"
	"github.com/silence-operator/silence-operator/pkg/config"
	"github.com/silence-operator/silence-operator/pkg/operator"
	"os"
)

func main() {
	path, exists := os.LookupEnv("SILENCE_OPERATOR_CONFIG_FILE")
	if !exists {
		log.Fatal().Msg("SILENCE_OPERATOR_CONFIG_FILE env variable is not declared")
	}

	conf, err := config.LoadConfig(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().Any("configuration", conf).Msg("Loaded configuration")

	operator.Run(conf)

}
