package kind

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	kindapp "sigs.k8s.io/kind/cmd/kind/app"
	"sigs.k8s.io/kind/pkg/cmd"
	"strings"
)

func Run() {
	suffix := uuid.NewString()[:6]
	defer deleteKindCLuster(suffix)
	createKindCLuster(suffix)
	log.Info().Str("cluster_name", suffix).Msg("Here I can do all the funny staff in kind cluster")
}

func createKindCLuster(name string) {
	args := []string{"create", "cluster", "--name", name}
	if err := kindapp.Run(cmd.NewLogger(), cmd.StandardIOStreams(), args); err != nil {
		log.Fatal().Err(err).Any("args", strings.Join(args, " ")).Msg("Failed to create kind cluster")
	}
}

func deleteKindCLuster(name string) {
	args := []string{"delete", "cluster", "--name", name}
	if err := kindapp.Run(cmd.NewLogger(), cmd.StandardIOStreams(), args); err != nil {
		log.Error().Err(err).Any("args", strings.Join(args, " ")).Msg("There was a problem deleting kind cluster")
	}
}
