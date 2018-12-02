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
	var dryRun bool
	flag.BoolVar(&dryRun, "dryRun", false, "enables dry run mode that evaluates all strategies but does not actualyl perform kicking (optional)")
	flag.Parse()

	engine.Exec(kickerConfPath, dryRun)
}
