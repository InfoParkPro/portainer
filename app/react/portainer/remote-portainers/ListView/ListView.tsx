import { Edit, Layers, PlugZap, Plus, Trash2 } from 'lucide-react';
import { useRouter } from '@uirouter/react';

import {
  notifyError,
  notifySuccess,
} from '@/portainer/services/notifications';
import { isoDateFromTimestamp } from '@/portainer/filters/filters';

import { PageHeader } from '@@/PageHeader';
import { Button, LoadingButton } from '@@/buttons';
import { Widget } from '@@/Widget';
import { confirmDelete } from '@@/modals/confirm';

import {
  useDeleteRemotePortainerMutation,
  useRemotePortainers,
  useTestRemotePortainerMutation,
} from '../queries';
import { RemotePortainer } from '../types';

export function ListView() {
  const router = useRouter();
  const remotePortainersQuery = useRemotePortainers();
  const deleteMutation = useDeleteRemotePortainerMutation();
  const testMutation = useTestRemotePortainerMutation();

  const remotePortainers = remotePortainersQuery.data || [];

  return (
    <>
      <PageHeader
        title="Remote Portainers"
        breadcrumbs="Remote Portainer management"
        reload
      />

      <Widget>
        <Widget.Title icon={PlugZap} title="Remote Portainers" />
        <Widget.Body loading={remotePortainersQuery.isLoading}>
          <div className="mb-3 flex justify-end">
            <Button
              icon={Plus}
              onClick={() => router.stateService.go('portainer.remotePortainers.new')}
              data-cy="remote-portainer-add-button"
            >
              Add remote Portainer
            </Button>
          </div>

          <div className="table-responsive">
            <table className="table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>URL</th>
                  <th>TLS skip verify</th>
                  <th>Updated</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {remotePortainers.map((remotePortainer) => (
                  <tr key={remotePortainer.Id}>
                    <td>{remotePortainer.Name}</td>
                    <td>{remotePortainer.URL}</td>
                    <td>{remotePortainer.TLSSkipVerify ? 'Yes' : 'No'}</td>
                    <td>
                      {remotePortainer.UpdatedAt
                        ? isoDateFromTimestamp(remotePortainer.UpdatedAt)
                        : '-'}
                    </td>
                    <td className="text-right">
                      <div className="inline-flex gap-1">
                        <LoadingButton
                          color="default"
                          icon={PlugZap}
                          data-cy={`remote-portainer-test-${remotePortainer.Id}`}
                          isLoading={
                            testMutation.isLoading &&
                            testMutation.variables === remotePortainer.Id
                          }
                          loadingText="Testing"
                          onClick={() => handleTest(remotePortainer)}
                        >
                          Test
                        </LoadingButton>
                        <Button
                          color="default"
                          icon={Layers}
                          data-cy={`remote-portainer-stacks-${remotePortainer.Id}`}
                          onClick={() =>
                            router.stateService.go(
                              'portainer.remotePortainers.stacks',
                              { id: remotePortainer.Id }
                            )
                          }
                        >
                          Stacks
                        </Button>
                        <Button
                          color="default"
                          icon={Edit}
                          data-cy={`remote-portainer-edit-${remotePortainer.Id}`}
                          onClick={() =>
                            router.stateService.go(
                              'portainer.remotePortainers.edit',
                              { id: remotePortainer.Id }
                            )
                          }
                        >
                          Edit
                        </Button>
                        <Button
                          color="dangerlight"
                          icon={Trash2}
                          data-cy={`remote-portainer-delete-${remotePortainer.Id}`}
                          onClick={() => handleDelete(remotePortainer)}
                        >
                          Delete
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
                {remotePortainers.length === 0 && !remotePortainersQuery.isLoading && (
                  <tr>
                    <td colSpan={5} className="text-center text-muted">
                      No remote Portainers configured.
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

  async function handleTest(remotePortainer: RemotePortainer) {
    try {
      const status = await testMutation.mutateAsync(remotePortainer.Id);
      notifySuccess(
        'Remote Portainer connection successful',
        `${remotePortainer.Name} ${status.Version}`
      );
    } catch (err) {
      notifyError(
        'Failure',
        err as Error,
        `Unable to connect to ${remotePortainer.Name}`
      );
    }
  }

  async function handleDelete(remotePortainer: RemotePortainer) {
    const confirmed = await confirmDelete(
      `Remove remote Portainer "${remotePortainer.Name}"?`
    );
    if (!confirmed) {
      return;
    }

    deleteMutation.mutate(remotePortainer.Id, {
      onSuccess() {
        notifySuccess('Remote Portainer removed', remotePortainer.Name);
      },
      onError(err) {
        notifyError(
          'Failure',
          err as Error,
          `Unable to remove ${remotePortainer.Name}`
        );
      },
    });
  }
}
