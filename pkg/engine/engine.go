package engine

import (
	"log"
	"time"

	"github.com/curlymon/kicker/pkg/client"
	"github.com/curlymon/kicker/pkg/conf"
	"github.com/curlymon/kicker/pkg/strategy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // this loads the gcp plugin (only required to authenticate against GKE clusters).
)

// Exec runs the program using the config defined at the kickerConfPath. If kickerConfPath is left empty it will attempt
// load from the environment.
func Exec(kickerConfPath string) {
	config, err := conf.LoadConf(kickerConfPath)
	if err != nil {
		log.Fatalln(err)
	}

	interval := time.Duration(config.CheckInterval) * time.Second

	clientset, err := client.New(config)
	if err != nil {
		log.Fatalln(err)
	}

	strats, err := strategy.NewGroup(config.Criteria)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		podList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln(err)
		}

		pods := podList.Items

		log.Printf("There are %d pods in the cluster\n", len(pods))
		log.Printf("Running %d strategies...\n", len(strats))

		for _, strat := range strats {
			sc := strat.Criteria()
			log.Printf("running %s strategy...", sc.Name)
			toKill := strat.Evaluate(pods)
			for _, pod := range toKill {
				log.Printf("kicking: %s...\n", pod.Name)
				criteria := strat.Criteria()
				fore := metav1.DeletePropagationForeground
				opts := &metav1.DeleteOptions{
					PropagationPolicy:  &fore,
					GracePeriodSeconds: &criteria.GracePeriod,
				}
				err := clientset.CoreV1().Pods(criteria.Namespace).Delete(pod.Name, opts)
				if err != nil {
					log.Printf("error kicking pod '%s': %s", pod.Name, err)
				}
			}

			// let's not evaluate the same pods for other strategies
			pods = strategy.FilterPodSet(toKill)(pods)

			log.Printf("completed %s strategy", sc.Name)
		}

		log.Printf("sleeping for %s", interval)
		time.Sleep(interval)
	}
}
