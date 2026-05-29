import { FileText, Info, Terminal, Trash2 } from 'lucide-react';

import { removeContainer } from '@/react/docker/containers/containers.service';
import { Authorized } from '@/react/hooks/useUser';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import {
  notifyError,
  notifySuccess,
} from '@/portainer/services/notifications';

import { Icon } from '@@/Icon';
import { Link } from '@@/Link';
import { confirmDelete } from '@@/modals/confirm';

interface State {
  showQuickActionInspect: boolean;
  showQuickActionLogs: boolean;
}

export function TaskTableQuickActions({
  taskId,
  containerId,
  nodeName,
  state = {
    showQuickActionInspect: true,
    showQuickActionLogs: true,
  },
}: {
  taskId: string;
  containerId?: string;
  nodeName?: string;
  state?: State;
}) {
  return (
    <div className="inline-flex space-x-1">
      {state.showQuickActionLogs && (
        <Authorized authorizations="DockerTaskLogs">
          <Link
            to="docker.tasks.task.logs"
            params={{ id: taskId }}
            title="Logs"
            data-cy="docker-task-logs-link"
          >
            <Icon icon={FileText} className="space-right" />
          </Link>
        </Authorized>
      )}

      {state.showQuickActionInspect && (
        <Authorized authorizations="DockerTaskInspect">
          <Link
            to="docker.tasks.task"
            params={{ id: taskId }}
            title="Inspect"
            data-cy="docker-task-inspect-link"
          >
            <Icon icon={Info} className="space-right" />
          </Link>
        </Authorized>
      )}

      <TaskContainerQuickActions containerId={containerId} nodeName={nodeName} />
    </div>
  );
}

export function TaskContainerQuickActions({
  containerId,
  nodeName,
}: {
  containerId?: string;
  nodeName?: string;
}) {
  const environmentId = useEnvironmentId();

  if (!containerId) {
    return null;
  }

  const taskContainerId = containerId;

  return (
    <>
      <Authorized authorizations="DockerExecStart">
        <Link
          to="docker.containers.container.exec"
          params={{ id: taskContainerId, nodeName }}
          title="Exec Console"
          data-cy={`container-exec-${taskContainerId}`}
        >
          <Icon icon={Terminal} className="space-right" />
        </Link>
      </Authorized>

      <Authorized authorizations="DockerContainerDelete">
        <button
          type="button"
          className="border-0 bg-transparent p-0"
          title="Force remove container"
          data-cy={`container-delete-${taskContainerId}`}
          onClick={handleRemove}
        >
          <Icon icon={Trash2} className="space-right" />
        </button>
      </Authorized>
    </>
  );

  async function handleRemove() {
    const confirmed = await confirmDelete(
      'Do you want to force remove this task container?'
    );

    if (!confirmed) {
      return;
    }

    try {
      await removeContainer(environmentId, taskContainerId, {
        nodeName,
        removeVolumes: false,
      });
      notifySuccess('Success', 'Container successfully removed');
    } catch (err) {
      notifyError('Failure', err as Error, 'Unable to remove container');
    }
  }
}
