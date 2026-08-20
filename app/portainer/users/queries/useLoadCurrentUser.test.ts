import { beforeEach, describe, expect, it, vi } from 'vitest';

import axios from '@/react/portainer/services/axios/axios';

import { getCurrentUser } from './useLoadCurrentUser';

vi.mock('@/react/portainer/services/axios/axios', () => ({
  default: {
    get: vi.fn(),
  },
}));

vi.mock('@/react-tools/react-query', () => ({
  withError: () => ({}),
}));

const mockedAxios = vi.mocked(axios);

function unauthorizedError() {
  return {
    response: {
      status: 401,
    },
  };
}

describe('getCurrentUser', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('retries once on an immediate 401 when requested by the login flow', async () => {
    const user = { Id: 1, Username: 'admin' };

    mockedAxios.get
      .mockRejectedValueOnce(unauthorizedError())
      .mockResolvedValueOnce({ data: user });

    await expect(
      getCurrentUser({ retryUnauthorizedOnce: true, retryDelay: 0 })
    ).resolves.toBe(user);

    expect(mockedAxios.get).toHaveBeenCalledTimes(2);
    expect(mockedAxios.get).toHaveBeenNthCalledWith(1, '/users/me');
    expect(mockedAxios.get).toHaveBeenNthCalledWith(2, '/users/me');
  });

  it('does not retry a 401 by default', async () => {
    const error = unauthorizedError();

    mockedAxios.get.mockRejectedValueOnce(error);

    await expect(getCurrentUser()).rejects.toBe(error);

    expect(mockedAxios.get).toHaveBeenCalledTimes(1);
  });

  it('does not retry non-401 errors', async () => {
    const error = {
      response: {
        status: 500,
      },
    };

    mockedAxios.get.mockRejectedValueOnce(error);

    await expect(
      getCurrentUser({ retryUnauthorizedOnce: true, retryDelay: 0 })
    ).rejects.toBe(error);

    expect(mockedAxios.get).toHaveBeenCalledTimes(1);
  });
});
