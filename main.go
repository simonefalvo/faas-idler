package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/smvfal/faas-idler/pkg/scaling"
	"github.com/smvfal/faas-monitor/pkg/metrics"
	"github.com/smvfal/faas-monitor/pkg/metrics/prometheus"
)

func main() {

	log.Print("faas-idler started")

	env, ok := os.LookupEnv("INACTIVITY_DURATION")
	if !ok {
		log.Fatal("$INACTIVITY_DURATION not set")
	}
	val, err := strconv.Atoi(env)
	if err != nil {
		log.Fatal(err.Error())
	}
	inactivityDuration := int64(val)
	log.Printf("inactivity duration: %v minutes", inactivityDuration/60)

	env, ok = os.LookupEnv("RECONCILE_INTERVAL")
	if !ok {
		log.Fatal("$RECONCILE_INTERVAL not set")
	}
	val, err = strconv.Atoi(env)
	if err != nil {
		log.Fatal(err.Error())
	}
	reconcileInterval := time.Duration(val)
	log.Printf("reconcile interval: %v seconds", val)

	var metricsProvider metrics.Provider
	metricsProvider = &metrics.FaasProvider{}

	for {

		// sleep
		time.Sleep(reconcileInterval * time.Second)

		// retrieve functions
		log.Print("retrieving scalable functions...")
		functions, err := scaling.ScalableFunctions()
		if err != nil {
			log.Printf("ERROR: %s", err)
			continue
		}
		if len(functions) == 0 {
			log.Printf("not found scalable functions")
			continue
		}
		log.Printf("scalable functions: %v", functions)

		// scale idle functions to zero
		log.Print("checking idle functions...")
		for _, function := range functions {
			_, err := metricsProvider.FunctionInvocationRate(function, inactivityDuration)
			if err != nil {
				if _, ok := err.(*prometheus.IdleError); ok {
					log.Printf("found an idle function: %s", function)
					err := scaling.ScaleFunction(0, function)
					if err != nil {
						log.Printf("ERROR: %s", err)
						continue
					}
					log.Printf("%s scaled down to zero", function)
				} else {
					log.Printf("ERROR: %s", err)
					continue
				}
			} else {
				log.Printf("not found idle functions")
			}
		}
	}
}
