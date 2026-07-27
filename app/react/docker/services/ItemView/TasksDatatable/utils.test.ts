import { describe, expect, it } from 'vitest';

import { ContainerStatus } from '@/react/docker/containers/types';

import { getTaskDisplayStatus } from './utils';
import { DecoratedTask } from './types';

function task(
  taskState: string,
  containerStatus?: ContainerStatus
): DecoratedTask {
  return {
    Status: { State: taskState },
    Container: containerStatus ? { Status: containerStatus } : undefined,
  } as DecoratedTask;
}

describe('getTaskDisplayStatus', () => {
  it('uses container health status when a swarm task has a health check', () => {
    expect(
      getTaskDisplayStatus(task('running', ContainerStatus.Unhealthy))
    ).toBe(ContainerStatus.Unhealthy);
    expect(getTaskDisplayStatus(task('running', ContainerStatus.Healthy))).toBe(
      ContainerStatus.Healthy
    );
    expect(
      getTaskDisplayStatus(task('running', ContainerStatus.Starting))
    ).toBe(ContainerStatus.Starting);
  });

  it('keeps task state when the container has no health status', () => {
    expect(getTaskDisplayStatus(task('running', ContainerStatus.Running))).toBe(
      'running'
    );
    expect(getTaskDisplayStatus(task('failed'))).toBe('failed');
  });
});
