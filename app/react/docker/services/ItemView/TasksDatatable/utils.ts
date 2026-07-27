import { ContainerStatus } from '@/react/docker/containers/types';

import { DecoratedTask } from './types';

const healthStatuses = [
  ContainerStatus.Healthy,
  ContainerStatus.Unhealthy,
  ContainerStatus.Starting,
];

export function getTaskDisplayStatus(task: DecoratedTask) {
  const containerStatus = task.Container?.Status;

  if (containerStatus && healthStatuses.includes(containerStatus)) {
    return containerStatus;
  }

  return task.Status?.State;
}

export function getTaskDisplayStatusBadge(status?: string) {
  if (!status) {
    return undefined;
  }

  switch (status) {
    case ContainerStatus.Healthy:
    case ContainerStatus.Running:
      return 'success';
    case ContainerStatus.Unhealthy:
    case ContainerStatus.Starting:
      return 'warning';
    default:
      return undefined;
  }
}
