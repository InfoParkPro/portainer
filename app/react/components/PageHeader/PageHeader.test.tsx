import { render } from '@testing-library/react';

import { UserViewModel } from '@/portainer/models/user';
import { withTestRouter } from '@/react/test-utils/withRouter';
import { withUserProvider } from '@/react/test-utils/withUserProvider';
import { withTestQueryProvider } from '@/react/test-utils/withTestQuery';

import { PageHeader } from './PageHeader';
import { setBrowserTitleEnvironment } from './browser-title';

beforeEach(() => {
  setBrowserTitleEnvironment('docker3');
});

test('should display a PageHeader', async () => {
  const username = 'username';
  const user = new UserViewModel({ Username: username });

  const Wrapped = withTestQueryProvider(
    withUserProvider(withTestRouter(PageHeader), user)
  );

  const title = 'title';
  const { queryByText } = render(<Wrapped title={title} />);

  const heading = queryByText(title);
  expect(heading).toBeVisible();

  expect(queryByText(username)).toBeVisible();
});

test('should update browser title with page context and environment name', async () => {
  const Wrapped = withTestQueryProvider(withTestRouter(PageHeader));

  render(
    <Wrapped
      title="Stack details"
      breadcrumbs={[{ label: 'Stacks', link: '^' }, 'redis']}
    />
  );

  expect(document.title).toBe('redis - Stack details - docker3');
});

test('should not duplicate equal browser title parts', async () => {
  const Wrapped = withTestQueryProvider(withTestRouter(PageHeader));

  render(<Wrapped title="Stacks" breadcrumbs="Stacks" />);

  expect(document.title).toBe('Stacks - docker3');
});
