package main

import (
	"assignment/lib/shared_lib/dto"
	logs "assignment/lib/shared_lib/logs"
	"assignment_service/pkg/app"
	"assignment_service/pkg/infra/cache"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	env "assignment/lib/shared_lib/env"
	"assignment/lib/shared_lib/trace"
	"assignment/lib/shared_lib/transport"
	network "assignment_service/pkg/infra/network"

	log "github.com/sirupsen/logrus"
)

var (
	Version string // set by build script
)

func init() {
	logs.Init()
}

func main() {
	queueConnectionStr := env.WithDefaultString("ASSIGNMENT_SERVICE_QUEUE_CONN", "amqp://guest:guest@127.0.0.1:5672/")
	serviceName := env.WithDefaultString("ASSIGNMENT_SERVICE_NAME", "assignment-service")
	queueName := env.WithDefaultString("ASSIGNMENT_SERVICE_QUEUE_NAME", "assignmentQueue")

	log.Info(fmt.Sprintf("Connected %s version %s", serviceName, Version))

	cancelCtx, cancelFtn := context.WithCancel(context.Background())
	traceShutdown := trace.Init(cancelCtx, map[string]string{}, serviceName, Version)

	orderedMap := cache.NewOrderedMap[string, string]()
	cmdUsecase := app.NewUsecase(orderedMap)
	dtoAdapter := dto.NewDTOAdapter()
	err := dtoAdapter.RegisterCommandTypeValidation()
	if err != nil {
		log.Fatal("register dto validation failed: ", err)
	}

	connector := transport.NewMQConnector()
	subscriber := network.NewSubscriber(queueConnectionStr, queueName, dtoAdapter, cmdUsecase, connector)
	if err := subscriber.Connect(cancelCtx); err != nil {
		log.Fatal("Connect subscriber failed: ", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	cancelFtn()
	subscriber.Close()
	traceShutdown()
	log.Trace(fmt.Sprintf("stopped %s version %s", serviceName, Version))
}
