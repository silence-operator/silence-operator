package main

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"os"
	kind "sigs.k8s.io/kind/cmd/kind/app"
)

func main() {
	suffix := uuid.NewString()[:6]
	defer deleteKindCLuster(suffix)
	createKindCLuster(suffix)
	log.Info().Str("cluster_name", suffix).Msg("Here I can do all the funny staff in kind cluster")
}

func createKindCLuster(name string) {
	os.Args = []string{"cmd", "create", "cluster", "--name", name}
	kind.Main()
}
func deleteKindCLuster(name string) {
	os.Args = []string{"cmd", "delete", "cluster", "--name", name}
	kind.Main()
}
