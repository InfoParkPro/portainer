import { useQuery } from '@tanstack/react-query';

import axios from '@/react/portainer/services/axios/axios';
import { withError } from '@/react-tools/react-query';

import { buildUrl } from '../user.service';
import { User } from '../types';

import { userQueryKeys } from './queryKeys';

export interface CurrentUserResponse extends User {
  forceChangePassword?: boolean;
}

export function useLoadCurrentUser({ staleTime }: { staleTime?: number } = {}) {
  return useQuery(userQueryKeys.me(), () => getCurrentUser(), {
    ...withError('Unable to retrieve user details'),
    staleTime,
  });
}

type GetCurrentUserOptions = {
  retryUnauthorizedOnce?: boolean;
  retryDelay?: number;
};

export async function getCurrentUser(options: GetCurrentUserOptions = {}) {
  try {
    return await fetchCurrentUser();
  } catch (err) {
    if (!options.retryUnauthorizedOnce || !isUnauthorizedError(err)) {
      throw err;
    }

    await delay(options.retryDelay ?? 300);

    return fetchCurrentUser();
  }
}

async function fetchCurrentUser() {
  const { data: user } = await axios.get<CurrentUserResponse>(
    buildUrl(undefined, 'me')
  );

  return user;
}

function isUnauthorizedError(err: unknown) {
  return (
    typeof err === 'object' &&
    err !== null &&
    'response' in err &&
    (err as { response?: { status?: number } }).response?.status === 401
  );
}

function delay(ms: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}
