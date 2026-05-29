import { CellContext } from '@tanstack/react-table';

import { ContainerQuickActions } from '@/react/docker/containers/components/ContainerQuickActions';
import { useCurrentEnvironment } from '@/react/hooks/useCurrentEnvironment';
import { isAgentEnvironment } from '@/react/portainer/environments/utils';
import { QuickActionsState } from '@/react/docker/containers/components/ContainerQuickActions/ContainerQuickActions';
import {
  TaskContainerQuickActions,
  TaskTableQuickActions,
} from '@/react/docker/services/common/TaskTableQuickActions';

import { DecoratedTask } from '../types';

import { columnHelper } from './helper';

export const actions = columnHelper.display({
  header: 'Actions',
  cell: Cell,
});

function Cell({
  row: { original: item },
}: CellContext<DecoratedTask, unknown>) {
  const environmentQuery = useCurrentEnvironment();

  if (!environmentQuery.data) {
    return null;
  }
  const containerId = item.ContainerId || item.Container?.Id;
  const nodeName = item.Container?.NodeName;

  const state: QuickActionsState = {
    showQuickActionAttach: false,
    showQuickActionExec: !containerId,
    showQuickActionInspect: true,
    showQuickActionLogs: true,
    showQuickActionStats: true,
  };
  const isAgent = isAgentEnvironment(environmentQuery.data.Type);

  return isAgent && item.Container ? (
    <>
      <ContainerQuickActions
        containerId={item.Container.Id}
        nodeName={item.Container.NodeName}
        status={item.Container.Status}
        state={state}
      />
      <TaskContainerQuickActions containerId={containerId} nodeName={nodeName} />
    </>
  ) : (
    <TaskTableQuickActions
      taskId={item.Id}
      containerId={containerId}
      nodeName={nodeName}
    />
  );
}
