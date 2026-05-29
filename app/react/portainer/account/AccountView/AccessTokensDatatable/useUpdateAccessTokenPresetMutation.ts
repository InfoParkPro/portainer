import { useMutation, useQueryClient } from '@tanstack/react-query';

import { withError, withInvalidate } from '@/react-tools/react-query';
import { useCurrentUser } from '@/react/hooks/useUser';
import axios, { parseAxiosError } from '@/portainer/services/axios/axios';

import {
  AccessToken,
  AccessTokenAccessPreset,
} from '../../access-tokens/types';
import { buildUrl } from '../../access-tokens/queries/build-url';
import { queryKeys } from '../../access-tokens/queries/query-keys';

interface UpdateAccessTokenPresetPayload {
  tokenId: AccessToken['id'];
  accessPreset: AccessTokenAccessPreset;
}

export function useUpdateAccessTokenPresetMutation() {
  const { user } = useCurrentUser();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ tokenId, accessPreset }: UpdateAccessTokenPresetPayload) =>
      updateAccessTokenPreset(user.Id, tokenId, accessPreset),
    ...withError('Failed to update access token'),
    ...withInvalidate(queryClient, [queryKeys.base(user.Id)]),
  });
}

async function updateAccessTokenPreset(
  userId: number,
  tokenId: AccessToken['id'],
  accessPreset: AccessTokenAccessPreset
) {
  try {
    await axios.put(buildUrl(userId, tokenId), { accessPreset });
  } catch (e) {
    throw parseAxiosError(e, 'Unable to update access token');
  }
}
