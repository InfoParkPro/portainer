import axios, { parseAxiosError } from '@/portainer/services/axios/axios';
import { StackId } from '@/react/common/stacks/types';

import {
  RemotePortainer,
  RemotePortainerId,
  RemotePortainerPayload,
  RemoteStack,
  RemoteStackFile,
  RemoteStackUpdatePayload,
  TestConnectionResponse,
} from './types';

const baseUrl = '/remote_portainers';

export async function getRemotePortainers() {
  try {
    const { data } = await axios.get<RemotePortainer[]>(baseUrl);
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve remote Portainers');
  }
}

export async function getRemotePortainer(id: RemotePortainerId) {
  try {
    const { data } = await axios.get<RemotePortainer>(`${baseUrl}/${id}`);
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve remote Portainer');
  }
}

export async function createRemotePortainer(payload: RemotePortainerPayload) {
  try {
    const { data } = await axios.post<RemotePortainer>(baseUrl, payload);
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to create remote Portainer');
  }
}

export async function updateRemotePortainer(
  id: RemotePortainerId,
  payload: RemotePortainerPayload
) {
  try {
    const { data } = await axios.put<RemotePortainer>(
      `${baseUrl}/${id}`,
      payload
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to update remote Portainer');
  }
}

export async function deleteRemotePortainer(id: RemotePortainerId) {
  try {
    await axios.delete(`${baseUrl}/${id}`);
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to delete remote Portainer');
  }
}

export async function testRemotePortainer(id: RemotePortainerId) {
  try {
    const { data } = await axios.post<TestConnectionResponse>(
      `${baseUrl}/${id}/test`
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to test remote Portainer');
  }
}

export async function getRemoteStacks(id: RemotePortainerId) {
  try {
    const { data } = await axios.get<RemoteStack[]>(`${baseUrl}/${id}/stacks`);
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve remote stacks');
  }
}

export async function getRemoteStackFile(
  id: RemotePortainerId,
  stackId: StackId
) {
  try {
    const { data } = await axios.get<RemoteStackFile>(
      `${baseUrl}/${id}/stacks/${stackId}/file`
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve remote stack file');
  }
}

export async function updateRemoteStack(
  id: RemotePortainerId,
  stackId: StackId,
  payload: RemoteStackUpdatePayload
) {
  try {
    const { data } = await axios.put<RemoteStack>(
      `${baseUrl}/${id}/stacks/${stackId}`,
      payload
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to update remote stack');
  }
}
