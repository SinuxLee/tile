package service

import "github.com/sinuxlee/tile/internal/store"

type UseCase interface {
	GetNodeID(path, addr, service string) (int, error)
}

func NewUseCase(d store.Dao) UseCase {
	return &useCaseImpl{
		dao: d,
	}
}

type useCaseImpl struct {
	dao store.Dao
}

func (c *useCaseImpl) GetNodeID(path, addr, service string) (int, error) {
	return c.dao.GetNodeID(path, addr, service)
}
