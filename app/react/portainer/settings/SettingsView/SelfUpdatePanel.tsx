import { useEffect, useState } from 'react';
import { RefreshCw } from 'lucide-react';

import {
  useSelfUpdatePlan,
  useStartSelfUpdateMutation,
} from '@/react/portainer/system/useSelfUpdate';
import { notifySuccess } from '@/portainer/services/notifications';

import { Widget } from '@@/Widget';
import { LoadingButton } from '@@/buttons';
import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { TextTip } from '@@/Tip/TextTip';
import { confirm } from '@@/modals/confirm';
import { buildConfirmButton } from '@@/modals/utils';
import { ModalType } from '@@/modals';

export function SelfUpdatePanel() {
  const planQuery = useSelfUpdatePlan();
  const mutation = useStartSelfUpdateMutation();
  const [targetImage, setTargetImage] = useState('');

  const plan = planQuery.data;

  useEffect(() => {
    if (plan && !targetImage) {
      setTargetImage(plan.targetImage);
    }
  }, [plan, targetImage]);

  if (!plan) {
    return null;
  }

  const canStart = plan.allowed && !!targetImage.trim();

  return (
    <Widget>
      <Widget.Title icon={RefreshCw} title="Portainer self-update" />
      <Widget.Body>
        <div className="form-horizontal">
          <div className="form-group">
            <div className="col-sm-12">
              {plan.allowed ? (
                <TextTip color="blue">
                  Starts a helper container that renames the current Portainer
                  container, stops it, and starts the replacement container.
                </TextTip>
              ) : (
                <TextTip color="orange">{plan.blockReason}</TextTip>
              )}
            </div>
          </div>

          <Detail label="Run mode" value={plan.mode} />
          <Detail label="Current container" value={plan.currentContainerName} />
          <Detail label="Current image" value={plan.currentImage} />
          <Detail label="Target container" value={plan.targetContainerName} />
          <Detail
            label="Rollback container"
            value={plan.rollbackContainerName}
          />

          <FormControl label="Target image" inputId="self_update_target_image">
            <Input
              id="self_update_target_image"
              value={targetImage}
              placeholder={plan.currentImage}
              onChange={(event) => setTargetImage(event.target.value)}
              disabled={!plan.allowed || mutation.isLoading}
              data-cy="settings-selfUpdateTargetImageInput"
            />
          </FormControl>

          <div className="form-group">
            <div className="col-sm-12">
              <LoadingButton
                type="button"
                icon={RefreshCw}
                isLoading={mutation.isLoading}
                disabled={!canStart}
                loadingText="Starting..."
                data-cy="settings-startSelfUpdateButton"
                onClick={handleStart}
              >
                Start self-update
              </LoadingButton>
            </div>
          </div>
        </div>
      </Widget.Body>
    </Widget>
  );

  async function handleStart() {
    const confirmed = await confirm({
      title: 'Start Portainer self-update?',
      message:
        'Portainer will start a helper container, stop the current container, and start the replacement container.',
      confirmButton: buildConfirmButton('Start update', 'danger'),
      modalType: ModalType.Warn,
    });

    if (!confirmed) {
      return;
    }

    mutation.mutate(
      { targetImage: targetImage.trim() },
      {
        onSuccess() {
          notifySuccess('Success', 'Self-update helper started');
        },
      }
    );
  }
}

function Detail({ label, value }: { label: string; value: string }) {
  if (!value) {
    return null;
  }

  return (
    <div className="form-group">
      <div className="col-sm-3 col-lg-2">
        <span className="control-label !block !pt-0 text-left font-medium">
          {label}
        </span>
      </div>
      <div className="col-sm-9 col-lg-10">
        <span className="break-all">{value}</span>
      </div>
    </div>
  );
}
