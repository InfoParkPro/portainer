import { EnvVar } from '@@/form-components/EnvironmentVariablesFieldset/types';

import { Stack } from '../../common/stacks/types';

export type RemotePortainerId = number;

export interface RemotePortainer {
  Id: RemotePortainerId;
  Name: string;
  URL: string;
  APIToken?: string;
  TLSSkipVerify: boolean;
  CreatedAt: number;
  UpdatedAt: number;
}

export interface RemotePortainerPayload {
  Name: string;
  URL: string;
  APIToken?: string;
  TLSSkipVerify: boolean;
}

export interface TestConnectionResponse {
  Status: string;
  Version: string;
}

export interface RemoteStackFile {
  StackFileContent: string;
}

export interface RemoteStackUpdatePayload {
  StackFileContent: string;
  Env: EnvVar[];
  Prune: boolean;
  RepullImageAndRedeploy: boolean;
}

export type RemoteStack = Stack;
