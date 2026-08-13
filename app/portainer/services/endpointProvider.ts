import { ping } from '@/react/docker/proxy/queries/usePing';
import { environmentStore } from '@/react/hooks/current-environment-store';
import {
  Environment,
  EnvironmentType,
} from '@/react/portainer/environments/types';
import { setBrowserTitleEnvironment } from '@/react/components/PageHeader/browser-title';

interface State {
  currentEndpoint: Environment | null;
  pingInterval: NodeJS.Timeout | null;
}

/* @ngInject */
export function EndpointProvider() {
  const state: State = {
    currentEndpoint: null,
    pingInterval: null,
  };

  environmentStore.subscribe(
    (state) => state.environmentId,
    (environmentId) => {
      if (!environmentId) {
        setCurrentEndpoint(null);
      }
    }
  );

  return { endpointID, setCurrentEndpoint, currentEndpoint, clean };

  function endpointID() {
    return state.currentEndpoint?.Id;
  }

  function setCurrentEndpoint(endpoint: Environment | null) {
    state.currentEndpoint = endpoint;

    if (state.pingInterval) {
      clearInterval(state.pingInterval);
      state.pingInterval = null;
    }

    if (endpoint && endpoint.Type === EnvironmentType.EdgeAgentOnDocker) {
      state.pingInterval = setInterval(() => ping(endpoint.Id), 60 * 1000);
    }

    if (endpoint === null) {
      sessionStorage.setItem(
        'portainer.environmentId',
        JSON.stringify(undefined)
      );
    }

    setBrowserTitleEnvironment(endpoint?.Name);
  }

  function currentEndpoint() {
    return state.currentEndpoint;
  }

  function clean() {
    setCurrentEndpoint(null);
    environmentStore.getState().clear();
  }
}

export type EndpointProviderInterface = ReturnType<typeof EndpointProvider>;
