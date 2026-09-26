package useractivitylog

import (
	"errors"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "user_activity_logs"
)

// Service represents a service for managing user activity logs.
type Service struct {
	dataservices.BaseDataService[portainer.UserActivityLog, int]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portainer.UserActivityLog, int]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

// Log stores a new user activity log entry.
func (service *Service) Log(entry *portainer.UserActivityLog) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, any) {
			entry.ID = int(id)
			return int(id), *entry
		},
	)
}

// Logs returns all user activity log entries ordered by insertion.
func (service *Service) Logs() ([]portainer.UserActivityLog, error) {
	logs := make([]portainer.UserActivityLog, 0)

	err := service.Connection.GetAll(
		BucketName,
		&portainer.UserActivityLog{},
		func(obj any) (any, error) {
			entry, ok := obj.(*portainer.UserActivityLog)
			if !ok {
				return nil, errors.New("unable to cast element to user activity log")
			}

			logs = append(logs, *entry)
			return entry, nil
		},
	)

	return logs, err
}
