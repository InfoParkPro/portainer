package remoteportainer

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portainer.RemotePortainer, portainer.RemotePortainerID]
}

// GetNextIdentifier returns the next identifier for a remote Portainer connection.
func (service ServiceTx) GetNextIdentifier() int {
	return service.Tx.GetNextIdentifier(BucketName)
}

// Create creates a new remote Portainer connection.
func (service ServiceTx) Create(remotePortainer *portainer.RemotePortainer) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, any) {
			remotePortainer.ID = portainer.RemotePortainerID(id)
			return int(remotePortainer.ID), remotePortainer
		},
	)
}
