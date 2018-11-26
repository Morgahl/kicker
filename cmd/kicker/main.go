package main

import (
	"flag"

	"github.com/curlymon/kicker/pkg/engine"
	_ "github.com/curlymon/kicker/pkg/strategy/all" // this loads all strategies
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // this loads the gcp plugin (only required to authenticate against GKE clusters).
)

func main() {
	var kickerConfPath string
	flag.StringVar(&kickerConfPath, "config", "", "absolute path to the kicker config file (optional)")
	flag.Parse()

	engine.Exec(kickerConfPath)
}
