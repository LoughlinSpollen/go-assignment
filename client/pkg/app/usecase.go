package app

import (
	"assignment/lib/shared_lib/domain"
	"assignment_client/pkg/infra/network"
	"context"

	log "github.com/sirupsen/logrus"
)

type CmdUsecase interface {
	SendCommand(ctx context.Context, cmd *domain.Command) ([]domain.DataEntry, error)
}

type cmdUsecase struct {
	publisher network.Publisher
}

func NewCmdUseCase(publisher network.Publisher) CmdUsecase {
	log.Trace("NewCmdUseCase")
	return &cmdUsecase{
		publisher: publisher,
	}
}

func (c *cmdUsecase) SendCommand(ctx context.Context, cmd *domain.Command) ([]domain.DataEntry, error) {
	log.Trace("cmdUsecase SendCommand")

	return c.publisher.Publish(ctx, cmd)
}
