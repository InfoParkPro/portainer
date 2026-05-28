package remoteportainer

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "remote_portainers"

// Service represents a service for managing remote Portainer connections.
type Service struct {
	dataservices.BaseDataService[portainer.RemotePortainer, portainer.RemotePortainerID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portainer.RemotePortainer, portainer.RemotePortainerID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: service.BaseDataService.Tx(tx),
	}
}

// GetNextIdentifier returns the next identifier for a remote Portainer connection.
func (service *Service) GetNextIdentifier() int {
	return service.Connection.GetNextIdentifier(BucketName)
}

// Create creates a new remote Portainer connection.
func (service *Service) Create(remotePortainer *portainer.RemotePortainer) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, any) {
			remotePortainer.ID = portainer.RemotePortainerID(id)
			return int(remotePortainer.ID), remotePortainer
		},
	)
}
