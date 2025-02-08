package domain

import (
	"errors"
	"strings"
)

type CommandType int

const (
	AddItem CommandType = iota
	DeleteItem
	GetItem
	GetAllItems
)

func (s CommandType) String() string {
	types := []string{"ADD ITEM", "DELETE ITEM", "GET ITEM", "GET ALL ITEMS"}
	if s < AddItem || s > GetAllItems {
		return "unknown"
	}
	return types[s]
}

func ToCommandType(s string) (CommandType, error) {
	switch strings.ToUpper(s) {
	case "ADD ITEM":
		return AddItem, nil
	case "DELETE ITEM":
		return DeleteItem, nil
	case "GET ITEM":
		return GetItem, nil
	case "GET ALL ITEMS":
		return GetAllItems, nil
	default:
		return 0, errors.New("invalid CommandType")
	}
}

type Command struct {
	Name CommandType
	Data []DataEntry
}

type DataEntry struct {
	Key   string
	Value string
}
