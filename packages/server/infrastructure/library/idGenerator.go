package library

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/logger"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type IDGenerator interface {
	IDGenerate() (uuid.UUID, error)
	IDGenerateFromString(id string) (uuid.UUID, error)
}

type idGenerator struct {
	log *zap.Logger
}

func NewIDGenerator() IDGenerator {
	return &idGenerator{log: logger.New()}
}

func (g *idGenerator) IDGenerate() (uuid.UUID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		g.log.Error("ID Generator", zap.Any("error", err))
		return uuid.Nil, err
	}

	return id, nil
}

func (g *idGenerator) IDGenerateFromString(id string) (uuid.UUID, error) {
	uuID, err := uuid.FromString(id)
	if err != nil {
		g.log.Error("ID Generator From String", zap.Any("error", err))
		return uuid.Nil, err
	}

	return uuID, nil
}
