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
  temporaryAccessPreset?: AccessTokenAccessPreset | '';
  temporaryAccessExpiresAt?: number;
}

export function useUpdateAccessTokenPresetMutation() {
  const { user } = useCurrentUser();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      tokenId,
      accessPreset,
      temporaryAccessPreset,
      temporaryAccessExpiresAt,
    }: UpdateAccessTokenPresetPayload) =>
      updateAccessTokenPreset(user.Id, tokenId, {
        accessPreset,
        temporaryAccessPreset,
        temporaryAccessExpiresAt,
      }),
    ...withError('Failed to update access token'),
    ...withInvalidate(queryClient, [queryKeys.base(user.Id)]),
  });
}

async function updateAccessTokenPreset(
  userId: number,
  tokenId: AccessToken['id'],
  payload: Omit<UpdateAccessTokenPresetPayload, 'tokenId'>
) {
  try {
    await axios.put(buildUrl(userId, tokenId), payload);
  } catch (e) {
    throw parseAxiosError(e, 'Unable to update access token');
  }
}
