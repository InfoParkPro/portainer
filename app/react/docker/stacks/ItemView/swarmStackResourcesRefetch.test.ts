import { describe, expect, it, vi } from 'vitest';

import { refetchSwarmStackResourceQueries } from './swarmStackResourcesRefetch';

describe('refetchSwarmStackResourceQueries', () => {
  it('refetches active service and task queries directly', async () => {
    const servicesQuery = { refetch: vi.fn().mockResolvedValue(undefined) };
    const tasksQuery = { refetch: vi.fn().mockResolvedValue(undefined) };

    await refetchSwarmStackResourceQueries({
      servicesQuery,
      tasksQuery,
    });

    expect(servicesQuery.refetch).toHaveBeenCalledOnce();
    expect(tasksQuery.refetch).toHaveBeenCalledOnce();
  });

  it('refetches containers when the stack resources hook loaded them', async () => {
    const servicesQuery = { refetch: vi.fn().mockResolvedValue(undefined) };
    const tasksQuery = { refetch: vi.fn().mockResolvedValue(undefined) };
    const containersQuery = { refetch: vi.fn().mockResolvedValue(undefined) };

    await refetchSwarmStackResourceQueries({
      servicesQuery,
      tasksQuery,
      containersQuery,
    });

    expect(containersQuery.refetch).toHaveBeenCalledOnce();
  });
});
