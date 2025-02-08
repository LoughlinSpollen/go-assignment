package dto

import (
	"assignment/lib/shared_lib/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	validator "github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type CommandDTO struct {
	Name domain.CommandType `json:"name,omitempty" validate:"command_type_validation"`
	Data []DataEntryDTO     `json:"data,omitempty" validate:"dive"`
}

type DataEntryDTO struct {
	Key   string `json:"key,omitempty" validate:"max=250"`
	Value string `json:"value,omitempty" validate:"max=250"`
}

type ErrorDTO struct {
	Message string `json:"message"`
}

// Exported for testing purposes
type DtoAdapter struct {
	Validator *validator.Validate
}

func ValidateCommandType(fl validator.FieldLevel) bool {
	commandType, ok := fl.Field().Interface().(domain.CommandType)
	if !ok {
		return false
	}

	dto, ok := fl.Parent().Interface().(CommandDTO)
	if !ok {
		return false
	}

	switch commandType {
	case domain.AddItem:
		if len(dto.Data) != 1 {
			return false
		}
		for _, dataEntry := range dto.Data {
			if dataEntry.Key == "" || dataEntry.Value == "" {
				return false
			}
		}
	case domain.DeleteItem, domain.GetItem:
		if len(dto.Data) != 1 {
			return false
		}
		for _, dataEntry := range dto.Data {
			if dataEntry.Key == "" {
				return false
			}
		}
	case domain.GetAllItems:
		if len(dto.Data) != 0 {
			return false
		}
		return true
	default:
		return false
	}
	return true
}

type DTOAdapter interface {
	RegisterCommandTypeValidation() error
	ToDTO(ctx context.Context, cmd *domain.Command) ([]byte, error)
	FromDTO(ctx context.Context, cmdBytes []byte) (*domain.Command, error)
	ToErrorDTO(ctx context.Context, err error) ([]byte, error)
	FromErrorDTO(ctx context.Context, errBytes []byte) (error, error)
}

func NewDTOAdapter() DTOAdapter {
	log.Trace("DtoAdapter")

	return &DtoAdapter{
		Validator: validator.New(),
	}
}

func (a *DtoAdapter) RegisterCommandTypeValidation() error {
	err := a.Validator.RegisterValidation("command_type_validation", ValidateCommandType)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to register command type validation: %w", err)
		log.WithError(err).Error("failed to register command type validation")
		return wrappedErr
	}
	return nil
}

func (a *DtoAdapter) ToDTO(ctx context.Context, cmd *domain.Command) ([]byte, error) {
	log.Trace("DtoAdapter ToDTO")

	cmdDto := CommandDTO{
		Name: cmd.Name,
		Data: make([]DataEntryDTO, len(cmd.Data)),
	}
	for i, dataEntry := range cmd.Data {
		cmdDto.Data[i] = DataEntryDTO{
			Key:   dataEntry.Key,
			Value: dataEntry.Value,
		}
	}

	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	err := encoder.Encode(cmdDto)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("encoding failed")
		return nil, err
	}

	return buff.Bytes(), nil
}

func (a *DtoAdapter) FromDTO(ctx context.Context, cmdBytes []byte) (*domain.Command, error) {
	log.Trace("DtoAdapter FromDTO")

	var cmdDto CommandDTO
	decoder := json.NewDecoder(bytes.NewReader(cmdBytes))
	if err := decoder.Decode(&cmdDto); err != nil {
		log.WithContext(ctx).WithError(err).Warn("failed to decode command")
		return nil, err
	}

	if err := a.Validator.Struct(cmdDto); err != nil {
		log.WithContext(ctx).WithError(err).Warn("failed to validate command")
		return nil, err
	}

	cmd := domain.Command{
		Name: cmdDto.Name,
		Data: make([]domain.DataEntry, len(cmdDto.Data)),
	}
	for i, dataEntryDto := range cmdDto.Data {
		cmd.Data[i] = domain.DataEntry{
			Key:   dataEntryDto.Key,
			Value: dataEntryDto.Value,
		}
	}
	return &cmd, nil
}

func (a *DtoAdapter) ToErrorDTO(ctx context.Context, err error) ([]byte, error) {
	log.Trace("DtoAdapter ToErrorDTO")

	errDto := ErrorDTO{
		Message: err.Error(),
	}

	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	if err := encoder.Encode(errDto); err != nil {
		log.WithContext(ctx).WithError(err).Error("encoding failed")
		return nil, err
	}
	return buff.Bytes(), nil
}

func (a *DtoAdapter) FromErrorDTO(ctx context.Context, errBytes []byte) (error, error) {
	log.Trace("DtoAdapter FromErrorDTO")

	var errDto ErrorDTO
	decoder := json.NewDecoder(bytes.NewReader(errBytes))
	if err := decoder.Decode(&errDto); err != nil {
		log.WithContext(ctx).WithError(err).Warn("failed to decode error")
		return nil, err
	}
	return fmt.Errorf("%s", errDto.Message), nil
}
