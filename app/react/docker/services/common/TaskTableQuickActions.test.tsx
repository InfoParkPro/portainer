import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, test, vi } from 'vitest';

import { removeContainer } from '@/react/docker/containers/containers.service';

import { confirmDelete } from '@@/modals/confirm';


import { TaskTableQuickActions } from './TaskTableQuickActions';

vi.mock('@/react/hooks/useUser', () => ({
  Authorized: ({ children }: { children: React.ReactNode }) => children,
}));

vi.mock('@/react/hooks/useEnvironmentId', () => ({
  useEnvironmentId: () => 3,
}));

vi.mock('@/react/docker/containers/containers.service', () => ({
  removeContainer: vi.fn(),
}));

vi.mock('@/portainer/services/notifications', () => ({
  notifySuccess: vi.fn(),
  notifyError: vi.fn(),
}));

vi.mock('@@/modals/confirm', () => ({
  confirmDelete: vi.fn(),
}));

vi.mock('@@/Icon', () => ({
  Icon: ({ icon: Icon }: { icon: React.ComponentType }) => <Icon />,
}));

vi.mock('@@/Link', () => ({
  Link: ({
    children,
    'data-cy': dataCy,
    title,
  }: React.PropsWithChildren<{ 'data-cy'?: string; title?: string }>) => (
    <a href="/" data-cy={dataCy} title={title}>
      {children}
    </a>
  ),
}));

describe('TaskTableQuickActions', () => {
  test('renders container exec and delete actions when the task has a container id', () => {
    render(
      <TaskTableQuickActions
        taskId="task-1"
        containerId="container-1"
        nodeName="node-1"
      />
    );

    expect(screen.getByTitle('Exec Console')).toBeInTheDocument();
    expect(screen.getByTitle('Force remove container')).toBeInTheDocument();
  });

  test('force removes the task container after confirmation', async () => {
    vi.mocked(confirmDelete).mockResolvedValue(true);

    render(
      <TaskTableQuickActions
        taskId="task-1"
        containerId="container-1"
        nodeName="node-1"
      />
    );

    await userEvent.click(screen.getByTitle('Force remove container'));

    expect(confirmDelete).toHaveBeenCalledWith(
      'Do you want to force remove this task container?'
    );
    expect(removeContainer).toHaveBeenCalledWith(3, 'container-1', {
      nodeName: 'node-1',
      removeVolumes: false,
    });
  });
});
