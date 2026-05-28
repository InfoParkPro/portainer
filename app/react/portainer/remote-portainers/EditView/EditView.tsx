import { FormEvent, useEffect, useState } from 'react';
import { Save } from 'lucide-react';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';

import {
  notifyError,
  notifySuccess,
} from '@/portainer/services/notifications';

import { PageHeader } from '@@/PageHeader';
import { LoadingButton } from '@@/buttons';
import { Widget } from '@@/Widget';

import {
  useCreateRemotePortainerMutation,
  useRemotePortainer,
  useUpdateRemotePortainerMutation,
} from '../queries';
import { RemotePortainerPayload } from '../types';

const emptyForm: RemotePortainerPayload = {
  Name: '',
  URL: '',
  APIToken: '',
  TLSSkipVerify: false,
};

export function EditView() {
  const router = useRouter();
  const {
    params: { id },
  } = useCurrentStateAndParams();
  const remotePortainerId = id ? Number(id) : undefined;
  const isEdit = !!remotePortainerId;

  const remotePortainerQuery = useRemotePortainer(remotePortainerId);
  const createMutation = useCreateRemotePortainerMutation();
  const updateMutation = useUpdateRemotePortainerMutation(remotePortainerId || 0);

  const [formValues, setFormValues] =
    useState<RemotePortainerPayload>(emptyForm);

  useEffect(() => {
    if (remotePortainerQuery.data) {
      setFormValues({
        Name: remotePortainerQuery.data.Name,
        URL: remotePortainerQuery.data.URL,
        APIToken: '',
        TLSSkipVerify: remotePortainerQuery.data.TLSSkipVerify,
      });
    }
  }, [remotePortainerQuery.data]);

  return (
    <>
      <PageHeader
        title={isEdit ? 'Edit remote Portainer' : 'Add remote Portainer'}
        breadcrumbs="Remote Portainer management"
        reload
      />

      <Widget>
        <Widget.Title title="Connection details" />
        <Widget.Body loading={remotePortainerQuery.isLoading}>
          <form className="form-horizontal" onSubmit={handleSubmit}>
            <div className="form-group">
              <label
                className="col-sm-3 col-lg-2 control-label text-left"
                htmlFor="remote-portainer-name"
              >
                Name
              </label>
              <div className="col-sm-9 col-lg-10">
                <input
                  id="remote-portainer-name"
                  className="form-control"
                  value={formValues.Name}
                  onChange={(e) =>
                    setFormValues((values) => ({
                      ...values,
                      Name: e.target.value,
                    }))
                  }
                  required
                  data-cy="remote-portainer-name-input"
                />
              </div>
            </div>

            <div className="form-group">
              <label
                className="col-sm-3 col-lg-2 control-label text-left"
                htmlFor="remote-portainer-url"
              >
                URL
              </label>
              <div className="col-sm-9 col-lg-10">
                <input
                  id="remote-portainer-url"
                  className="form-control"
                  value={formValues.URL}
                  onChange={(e) =>
                    setFormValues((values) => ({
                      ...values,
                      URL: e.target.value,
                    }))
                  }
                  placeholder="https://remote.example:9443"
                  required
                  data-cy="remote-portainer-url-input"
                />
              </div>
            </div>

            <div className="form-group">
              <label
                className="col-sm-3 col-lg-2 control-label text-left"
                htmlFor="remote-portainer-token"
              >
                API token
              </label>
              <div className="col-sm-9 col-lg-10">
                <input
                  id="remote-portainer-token"
                  className="form-control"
                  type="password"
                  value={formValues.APIToken}
                  onChange={(e) =>
                    setFormValues((values) => ({
                      ...values,
                      APIToken: e.target.value,
                    }))
                  }
                  placeholder={
                    isEdit ? 'Leave empty to keep the current token' : ''
                  }
                  required={!isEdit}
                  data-cy="remote-portainer-token-input"
                />
              </div>
            </div>

            <div className="form-group">
              <div className="col-sm-offset-3 col-lg-offset-2 col-sm-9 col-lg-10">
                <label className="vertical-center">
                  <input
                    type="checkbox"
                    checked={formValues.TLSSkipVerify}
                    onChange={(e) =>
                      setFormValues((values) => ({
                        ...values,
                        TLSSkipVerify: e.target.checked,
                      }))
                    }
                  />{' '}
                  Skip TLS verification
                </label>
              </div>
            </div>

            <div className="form-group">
              <div className="col-sm-offset-3 col-lg-offset-2 col-sm-9 col-lg-10">
                <LoadingButton
                  icon={Save}
                  isLoading={createMutation.isLoading || updateMutation.isLoading}
                  loadingText="Saving"
                  data-cy="remote-portainer-save-button"
                >
                  Save
                </LoadingButton>
              </div>
            </div>
          </form>
        </Widget.Body>
      </Widget>
    </>
  );

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();

    try {
      if (isEdit && remotePortainerId) {
        await updateMutation.mutateAsync(formValues);
        notifySuccess('Remote Portainer updated', formValues.Name);
      } else {
        const created = await createMutation.mutateAsync(formValues);
        notifySuccess('Remote Portainer created', created.Name);
      }

      router.stateService.go('portainer.remotePortainers');
    } catch (err) {
      notifyError('Failure', err as Error, 'Unable to save remote Portainer');
    }
  }
}
