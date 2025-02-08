package main

import (
	"assignment/lib/shared_lib/domain"
	"assignment/lib/shared_lib/dto"
	env "assignment/lib/shared_lib/env"
	logs "assignment/lib/shared_lib/logs"
	"assignment/lib/shared_lib/trace"
	"assignment/lib/shared_lib/transport"
	"assignment_client/pkg/app"
	"assignment_client/pkg/infra/network"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	Version    string // set by build script
	clientName string
)

const (
	defaultClientName = "assignment"
)

func init() {
	logs.Init()
}

func main() {
	log.Infof("CLI version %s", Version)

	clientName = env.WithDefaultString("ASSIGNMENT_CLIENT_NAME", defaultClientName)
	rootCmd := &cobra.Command{
		Use:          "cli",
		Short:        fmt.Sprintf("%s CLI", clientName),
		SilenceUsage: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	traceShutdown := trace.Init(ctx, map[string]string{}, clientName, Version)
	queueConnectionStr := env.WithDefaultString("ASSIGNMENT_SERVICE_QUEUE_CONN", "amqp://guest:guest@127.0.0.1:5672/")
	serviceQueueName := env.WithDefaultString("ASSIGNMENT_SERVICE_QUEUE_NAME", "assignmentQueue")

	dtoAdapter := dto.NewDTOAdapter()
	dtoAdapter.RegisterCommandTypeValidation()
	connector := transport.NewMQConnector()
	publisher := network.NewPublisher(dtoAdapter, connector)
	err := publisher.Connect(ctx, serviceQueueName, queueConnectionStr)
	if err != nil {
		log.Fatal("Connect publisher failed: ", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit

		publisher.Close()
		traceShutdown()
		log.Trace(fmt.Sprintf("stopped %s", clientName))
		os.Exit(0)
	}()

	usecase := app.NewCmdUseCase(publisher)
	rootCmd.AddCommand(cliAddCmd(ctx, usecase))
	rootCmd.AddCommand(cliDeleteCmd(ctx, usecase))
	rootCmd.AddCommand(cliGetCmd(ctx, usecase))
	rootCmd.AddCommand(cliGetAllCmd(ctx, usecase))

	if err := rootCmd.Execute(); err != nil {
		log.Warn("Command failed: ", err)
	}
	quit <- syscall.SIGTERM

}

func cliAddCmd(ctx context.Context, usecase app.CmdUsecase) *cobra.Command {
	log.Trace("cliAddCmd")

	cliCmd := &cobra.Command{
		Use:   "add [KEY] [VALUE]",
		Short: "Add a key-value entry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			entries, err := usecase.SendCommand(ctx, &domain.Command{
				Name: domain.AddItem,
				Data: []domain.DataEntry{{Key: key, Value: value}},
			})
			if err != nil {
				return err
			}

			entry := entries[0]
			fmt.Printf("Successfully added %q = %q\n", entry.Key, entry.Value)
			return nil
		},
	}
	return cliCmd
}

func cliDeleteCmd(ctx context.Context, usecase app.CmdUsecase) *cobra.Command {
	log.Trace("cliDeleteCmd")

	cliCmd := &cobra.Command{
		Use:   "delete [KEY]",
		Short: "Delete an entry using a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			entries, err := usecase.SendCommand(ctx, &domain.Command{
				Name: domain.DeleteItem,
				Data: []domain.DataEntry{{Key: key}},
			})
			if err != nil {
				return err
			}

			entry := entries[0]
			fmt.Printf("Successfully deleted: %q = %q\n", entry.Key, entry.Value)
			return nil
		},
	}
	return cliCmd
}

func cliGetCmd(ctx context.Context, usecase app.CmdUsecase) *cobra.Command {
	log.Trace("cliGetCmd")
	cliCmd := &cobra.Command{
		Use:   "get [KEY]",
		Short: "Get one item by key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			entries, err := usecase.SendCommand(ctx, &domain.Command{
				Name: domain.GetItem,
				Data: []domain.DataEntry{{Key: key}},
			})
			if err != nil {
				return err
			}

			entry := entries[0]
			fmt.Printf("Successfully fetched  %q = %q\n", entry.Key, entry.Value)
			return nil
		},
	}
	return cliCmd
}

func cliGetAllCmd(ctx context.Context, usecase app.CmdUsecase) *cobra.Command {
	log.Trace("cliGetAllCmd")
	cliCmd := &cobra.Command{
		Use:   "getall",
		Short: "Get all key-value pairs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := usecase.SendCommand(ctx, &domain.Command{
				Name: domain.GetAllItems,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Successfully fetched %d key-value pairs.\n", len(entries))
			for _, entry := range entries {
				fmt.Printf("%q : %q\n", entry.Key, entry.Value)
			}
			return nil
		},
	}
	return cliCmd
}
