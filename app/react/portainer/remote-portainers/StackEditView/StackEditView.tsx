import { FormEvent, useEffect, useMemo, useState } from 'react';
import { Save } from 'lucide-react';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';

import {
  notifyError,
  notifySuccess,
} from '@/portainer/services/notifications';

import { PageHeader } from '@@/PageHeader';
import { LoadingButton } from '@@/buttons';
import { CodeEditor } from '@@/CodeEditor';
import { EnvironmentVariablesFieldset } from '@@/form-components/EnvironmentVariablesFieldset';
import { EnvVar } from '@@/form-components/EnvironmentVariablesFieldset/types';
import { Widget } from '@@/Widget';

import {
  useRemotePortainer,
  useRemoteStackFile,
  useRemoteStacks,
  useUpdateRemoteStackMutation,
} from '../queries';

export function StackEditView() {
  const router = useRouter();
  const {
    params: { id, stackId },
  } = useCurrentStateAndParams();
  const remotePortainerId = Number(id);
  const remoteStackId = Number(stackId);

  const remotePortainerQuery = useRemotePortainer(remotePortainerId);
  const stacksQuery = useRemoteStacks(remotePortainerId);
  const stackFileQuery = useRemoteStackFile(remotePortainerId, remoteStackId);
  const updateMutation = useUpdateRemoteStackMutation(
    remotePortainerId,
    remoteStackId
  );

  const stack = useMemo(
    () => stacksQuery.data?.find((item) => item.Id === remoteStackId),
    [remoteStackId, stacksQuery.data]
  );

  const [stackFileContent, setStackFileContent] = useState('');
  const [env, setEnv] = useState<EnvVar[]>([]);
  const [prune, setPrune] = useState(false);
  const [repullImageAndRedeploy, setRepullImageAndRedeploy] = useState(false);

  useEffect(() => {
    if (stackFileQuery.data) {
      setStackFileContent(stackFileQuery.data.StackFileContent);
    }
  }, [stackFileQuery.data]);

  useEffect(() => {
    if (stack?.Env) {
      setEnv(stack.Env);
    }
    if (stack?.Option) {
      setPrune(!!stack.Option.Prune);
    }
  }, [stack]);

  return (
    <>
      <PageHeader
        title={stack?.Name || 'Remote stack'}
        breadcrumbs={[
          { label: 'Remote Portainers', link: 'portainer.remotePortainers' },
          {
            label: remotePortainerQuery.data?.Name || 'Remote stacks',
            link: 'portainer.remotePortainers.stacks',
            linkParams: { id: remotePortainerId },
          },
          stack?.Name || 'Stack',
        ]}
        reload
      />

      <Widget>
        <Widget.Title title="Stack editor" />
        <Widget.Body
          loading={
            remotePortainerQuery.isLoading ||
            stacksQuery.isLoading ||
            stackFileQuery.isLoading
          }
        >
          <form onSubmit={handleSubmit}>
            <CodeEditor
              id="remote-stack-editor"
              type="yaml"
              value={stackFileContent}
              onChange={setStackFileContent}
              textTip="Edit the stack file stored on the remote Portainer"
              data-cy="remote-stack-editor"
            />

            <div className="mt-4">
              <EnvironmentVariablesFieldset values={env} onChange={setEnv} />
            </div>

            <div className="mt-4 flex flex-col gap-2">
              <label>
                <input
                  type="checkbox"
                  checked={prune}
                  onChange={(e) => setPrune(e.target.checked)}
                />{' '}
                Prune services that are no longer referenced
              </label>
              <label>
                <input
                  type="checkbox"
                  checked={repullImageAndRedeploy}
                  onChange={(e) => setRepullImageAndRedeploy(e.target.checked)}
                />{' '}
                Re-pull image and redeploy
              </label>
            </div>

            <div className="mt-4">
              <LoadingButton
                icon={Save}
                isLoading={updateMutation.isLoading}
                loadingText="Updating"
                disabled={!stackFileContent.trim()}
                data-cy="remote-stack-update-button"
              >
                Update remote stack
              </LoadingButton>
            </div>
          </form>
        </Widget.Body>
      </Widget>
    </>
  );

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();

    try {
      await updateMutation.mutateAsync({
        StackFileContent: stackFileContent,
        Env: env,
        Prune: prune,
        RepullImageAndRedeploy: repullImageAndRedeploy,
      });
      notifySuccess('Remote stack updated', stack?.Name || `${remoteStackId}`);
      router.stateService.go('portainer.remotePortainers.stacks', {
        id: remotePortainerId,
      });
    } catch (err) {
      notifyError('Failure', err as Error, 'Unable to update remote stack');
    }
  }
}
