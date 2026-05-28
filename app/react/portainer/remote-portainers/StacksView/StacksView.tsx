import { Edit } from 'lucide-react';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';

import { formatDate } from '@/portainer/filters/filters';
import { StackStatus, StackType } from '@/react/common/stacks/types';

import { PageHeader } from '@@/PageHeader';
import { Button } from '@@/buttons';
import { Widget } from '@@/Widget';

import { useRemotePortainer, useRemoteStacks } from '../queries';

export function StacksView() {
  const router = useRouter();
  const {
    params: { id },
  } = useCurrentStateAndParams();
  const remotePortainerId = Number(id);

  const remotePortainerQuery = useRemotePortainer(remotePortainerId);
  const stacksQuery = useRemoteStacks(remotePortainerId);
  const stacks = stacksQuery.data || [];

  return (
    <>
      <PageHeader
        title="Remote stacks"
        breadcrumbs={[
          { label: 'Remote Portainers', link: 'portainer.remotePortainers' },
          remotePortainerQuery.data?.Name || 'Remote stacks',
        ]}
        reload
      />

      <Widget>
        <Widget.Title title={remotePortainerQuery.data?.Name || 'Stacks'} />
        <Widget.Body loading={stacksQuery.isLoading || remotePortainerQuery.isLoading}>
          <div className="table-responsive">
            <table className="table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Type</th>
                  <th>Endpoint ID</th>
                  <th>Status</th>
                  <th>Updated</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {stacks.map((stack) => (
                  <tr key={stack.Id}>
                    <td>{stack.Name}</td>
                    <td>{stackTypeLabel(stack.Type)}</td>
                    <td>{stack.EndpointId}</td>
                    <td>{stackStatusLabel(stack.Status)}</td>
                    <td>{stack.UpdateDate ? formatDate(stack.UpdateDate) : '-'}</td>
                    <td className="text-right">
                      <Button
                        color="default"
                        icon={Edit}
                        data-cy={`remote-stack-edit-${stack.Id}`}
                        onClick={() =>
                          router.stateService.go(
                            'portainer.remotePortainers.stack',
                            {
                              id: remotePortainerId,
                              stackId: stack.Id,
                            }
                          )
                        }
                      >
                        Edit
                      </Button>
                    </td>
                  </tr>
                ))}
                {stacks.length === 0 && !stacksQuery.isLoading && (
                  <tr>
                    <td colSpan={6} className="text-center text-muted">
                      No stacks found on this remote Portainer.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </Widget.Body>
      </Widget>
    </>
  );
}

function stackTypeLabel(type: StackType) {
  switch (type) {
    case StackType.DockerSwarm:
      return 'Swarm';
    case StackType.DockerCompose:
      return 'Compose';
    case StackType.Kubernetes:
      return 'Kubernetes';
    default:
      return 'Unknown';
  }
}

function stackStatusLabel(status: StackStatus) {
  switch (status) {
    case StackStatus.Active:
      return 'Active';
    case StackStatus.Inactive:
      return 'Inactive';
    case StackStatus.Deploying:
      return 'Deploying';
    case StackStatus.Error:
      return 'Error';
    default:
      return 'Unknown';
  }
}
