package app

import (
	"assignment/lib/shared_lib/domain"
	"assignment_service/pkg/infra/cache"
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Usecase interface {
	HandleCommand(ctx context.Context, cmd *domain.Command) error
}

type usecase struct {
	cache cache.OrderedMap
}

func NewUsecase(orderedMap cache.OrderedMap) Usecase {
	log.Trace("new usecase")
	return &usecase{
		cache: orderedMap,
	}
}

func (u *usecase) HandleCommand(ctx context.Context, cmd *domain.Command) error {
	log.Trace("usecase HandleCommand")

	switch cmd.Name {
	case domain.AddItem:
		if ok := u.cache.Set(cmd.Data[0].Key, cmd.Data[0].Value); !ok {
			return fmt.Errorf("failed to set key: %s", cmd.Data[0].Key)
		}
	case domain.GetItem:
		value, ok := u.cache.Get(cmd.Data[0].Key)
		if !ok {
			return fmt.Errorf("failed to get key: %s", cmd.Data[0].Key)
		}
		cmd.Data[0].Value = value
	case domain.DeleteItem:
		if ok := u.cache.Delete(cmd.Data[0].Key); !ok {
			return fmt.Errorf("failed to delete key: %s", cmd.Data[0].Key)
		}
	case domain.GetAllItems:
		items := u.cache.Items()
		for _, item := range items {
			cmd.Data = append(cmd.Data, domain.DataEntry{Key: item.Key, Value: item.Value})
		}
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Name)
	}
	return nil
}
