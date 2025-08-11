package main

import (
	"github.com/Potterli20/golem/pkg/constants"
	"github.com/Potterli20/golem/pkg/logger"
	"github.com/Potterli20/golem/pkg/metrics"
	"github.com/Potterli20/golem/pkg/runner"
)

func main() {
	println("[Demo] Microservice example")

	logger.InitLogger(logger.Config{Level: constants.DebugLevel})
	r := runner.NewRunner()

	r.AddTask(metrics.NewTaskMetrics("/metrics", "9090", "demo"))

	// Now start all the tasks
	r.StartAndWait()
}
