import { useMutation, useQuery } from '@tanstack/react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios/axios';
import { isAxiosError } from '@/portainer/services/axios/utils/isAxiosError';
import { withError } from '@/react-tools/react-query';

import { buildUrl } from './build-url';
import { queryKeys } from './query-keys';

export interface SelfUpdatePlan {
  allowed: boolean;
  mode: string;
  blockReason: string;
  currentContainerId: string;
  currentContainerName: string;
  currentImage: string;
  targetImage: string;
  targetContainerName: string;
  rollbackContainerName: string;
}

export function useSelfUpdatePlan() {
  return useQuery(queryKeys.selfUpdatePlan(), getSelfUpdatePlan, {
    ...withError('Unable to retrieve self-update plan'),
  });
}

export function useStartSelfUpdateMutation() {
  return useMutation(startSelfUpdate, {
    ...withError('Unable to start self-update'),
  });
}

async function getSelfUpdatePlan() {
  try {
    const { data } = await axios.get<SelfUpdatePlan>(
      buildUrl('self-update/plan')
    );
    return data;
  } catch (error) {
    throw parseAxiosError(error as Error);
  }
}

async function startSelfUpdate({ targetImage }: { targetImage: string }) {
  try {
    await axios.post(buildUrl('self-update/start'), { targetImage });
  } catch (error) {
    if (!isAxiosError(error)) {
      throw error;
    }

    if (!error.response || !error.response.status) {
      return;
    }

    throw parseAxiosError(error);
  }
}
