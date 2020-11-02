package store

import "github.com/sinuxlee/tile/pkg/nid"

const (
	nodeIdRoot = "nodeId/"
)

type Dao interface {
	GetNodeID(path, addr, service string) (int, error)
}

func NewDao(named nid.NodeNamed) Dao {
	return &daoImpl{
		nodeNamed: named,
	}
}

type daoImpl struct {
	nodeNamed nid.NodeNamed
}

func (d *daoImpl) GetNodeID(path, addr, service string) (int, error) {
	return d.nodeNamed.GetNodeID(&nid.NameHolder{
		LocalPath:  path,
		LocalIP:    addr,
		ServiceKey: nodeIdRoot + service,
	})
}
