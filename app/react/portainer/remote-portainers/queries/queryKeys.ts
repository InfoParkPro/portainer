import { StackId } from '@/react/common/stacks/types';

import { RemotePortainerId } from '../types';

export const queryKeys = {
  base: () => ['remote-portainers'] as const,
  item: (id: RemotePortainerId) => [...queryKeys.base(), id] as const,
  stacks: (id: RemotePortainerId) =>
    [...queryKeys.item(id), 'stacks'] as const,
  stackFile: (id: RemotePortainerId, stackId: StackId) =>
    [...queryKeys.stacks(id), stackId, 'file'] as const,
};
