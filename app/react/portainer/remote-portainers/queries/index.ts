import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { StackId } from '@/react/common/stacks/types';
import { withError } from '@/react-tools/react-query';

import {
  createRemotePortainer,
  deleteRemotePortainer,
  getRemotePortainer,
  getRemotePortainers,
  getRemoteStackFile,
  getRemoteStacks,
  testRemotePortainer,
  updateRemotePortainer,
  updateRemoteStack,
} from '../remote-portainers.service';
import {
  RemotePortainerId,
  RemotePortainerPayload,
  RemoteStackUpdatePayload,
} from '../types';

import { queryKeys } from './queryKeys';

export function useRemotePortainers() {
  return useQuery(queryKeys.base(), getRemotePortainers, {
    ...withError('Unable to retrieve remote Portainers'),
  });
}

export function useRemotePortainer(id?: RemotePortainerId) {
  return useQuery(
    queryKeys.item(id || 0),
    () => getRemotePortainer(id as RemotePortainerId),
    {
      enabled: !!id,
      ...withError('Unable to retrieve remote Portainer'),
    }
  );
}

export function useRemoteStacks(id?: RemotePortainerId) {
  return useQuery(
    queryKeys.stacks(id || 0),
    () => getRemoteStacks(id as RemotePortainerId),
    {
      enabled: !!id,
      ...withError('Unable to retrieve remote stacks'),
    }
  );
}

export function useRemoteStackFile(
  id?: RemotePortainerId,
  stackId?: StackId
) {
  return useQuery(
    queryKeys.stackFile(id || 0, stackId || 0),
    () => getRemoteStackFile(id as RemotePortainerId, stackId as StackId),
    {
      enabled: !!id && !!stackId,
      ...withError('Unable to retrieve remote stack file'),
    }
  );
}

export function useCreateRemotePortainerMutation() {
  const queryClient = useQueryClient();

  return useMutation(createRemotePortainer, {
    onSuccess() {
      return queryClient.invalidateQueries(queryKeys.base());
    },
  });
}

export function useUpdateRemotePortainerMutation(id: RemotePortainerId) {
  const queryClient = useQueryClient();

  return useMutation(
    (payload: RemotePortainerPayload) => updateRemotePortainer(id, payload),
    {
      onSuccess() {
        queryClient.invalidateQueries(queryKeys.base());
        return queryClient.invalidateQueries(queryKeys.item(id));
      },
    }
  );
}

export function useDeleteRemotePortainerMutation() {
  const queryClient = useQueryClient();

  return useMutation(deleteRemotePortainer, {
    onSuccess() {
      return queryClient.invalidateQueries(queryKeys.base());
    },
  });
}

export function useTestRemotePortainerMutation() {
  return useMutation(testRemotePortainer);
}

export function useUpdateRemoteStackMutation(
  id: RemotePortainerId,
  stackId: StackId
) {
  const queryClient = useQueryClient();

  return useMutation(
    (payload: RemoteStackUpdatePayload) => updateRemoteStack(id, stackId, payload),
    {
      onSuccess() {
        queryClient.invalidateQueries(queryKeys.stacks(id));
        return queryClient.invalidateQueries(queryKeys.stackFile(id, stackId));
      },
    }
  );
}
