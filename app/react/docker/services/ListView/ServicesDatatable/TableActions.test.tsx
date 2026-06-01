import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { TableActions } from './TableActions';

vi.mock('@uirouter/react', () => ({
  useRouter: () => ({
    stateService: {
      reload: vi.fn(),
    },
  }),
}));

vi.mock('@/react/hooks/useEnvironmentId', () => ({
  useEnvironmentId: () => 1,
}));

vi.mock('@/react/hooks/useUser', () => ({
  Authorized: ({ children }: { children: React.ReactNode }) => children,
}));

vi.mock('@/portainer/services/notifications', () => ({
  notifySuccess: vi.fn(),
}));

vi.mock('./useRemoveServicesMutation', () => ({
  useRemoveServicesMutation: () => ({
    mutate: vi.fn(),
  }),
}));

vi.mock('./useForceUpdateServicesMutation', () => ({
  useForceUpdateServicesMutation: () => ({
    mutate: vi.fn(),
  }),
}));

test('renders manual refresh action when onRefresh is provided', async () => {
  const user = userEvent.setup();
  const onRefresh = vi.fn();

  render(
    <TableActions
      selectedItems={[]}
      isUpdateActionVisible
      onRefresh={onRefresh}
    />
  );

  await user.click(screen.getByRole('button', { name: 'Refresh' }));

  expect(onRefresh).toHaveBeenCalledOnce();
});

test('does not render manual refresh action when onRefresh is missing', () => {
  render(<TableActions selectedItems={[]} isUpdateActionVisible />);

  expect(
    screen.queryByRole('button', { name: 'Refresh' })
  ).not.toBeInTheDocument();
});
