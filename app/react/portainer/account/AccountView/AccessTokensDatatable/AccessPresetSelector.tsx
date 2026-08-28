import clsx from 'clsx';

import { Badge } from '@@/Badge';
import { Button } from '@@/buttons';
import { ButtonGroup } from '@@/buttons/ButtonGroup';

import {
  AccessToken,
  AccessTokenAccessPreset,
} from '../../access-tokens/types';

import { useUpdateAccessTokenPresetMutation } from './useUpdateAccessTokenPresetMutation';

const presets: Array<{
  value: AccessTokenAccessPreset;
  label: string;
  title: string;
}> = [
  {
    value: 'disabled',
    label: 'Off',
    title: 'Disable this access token',
  },
  {
    value: 'read_only',
    label: 'Read',
    title: 'Allow read-only API requests',
  },
  {
    value: 'power',
    label: 'Power',
    title: 'Allow read-only plus start, stop, restart, pause and unpause',
  },
  {
    value: 'manage',
    label: 'Manage',
    title: 'Allow full access permitted by the user account',
  },
];

const temporaryDurations = [
  { label: '10m', seconds: 10 * 60 },
  { label: '1h', seconds: 60 * 60 },
  { label: '4h', seconds: 4 * 60 * 60 },
];

interface Props {
  token: AccessToken;
}

export function AccessPresetSelector({ token }: Props) {
  const mutation = useUpdateAccessTokenPresetMutation();
  const currentPreset = token.accessPreset || 'manage';
  const now = Math.floor(Date.now() / 1000);
  const hasTemporaryAccess =
    token.temporaryAccessPreset &&
    token.temporaryAccessExpiresAt &&
    token.temporaryAccessExpiresAt > now;
  const canTemporarilyElevate =
    currentPreset !== 'disabled' && currentPreset !== 'manage';

  return (
    <div className="flex flex-wrap items-center gap-2">
      <ButtonGroup size="xsmall" aria-label="Access token preset">
        {presets.map((preset) => (
          <Button
            key={preset.value}
            color="light"
            size="xsmall"
            className={clsx('!static !z-auto', {
              active: currentPreset === preset.value,
            })}
            disabled={currentPreset === preset.value || mutation.isLoading}
            onClick={() =>
              mutation.mutate({
                tokenId: token.id,
                accessPreset: preset.value,
                temporaryAccessPreset: '',
                temporaryAccessExpiresAt: 0,
              })
            }
            title={preset.title}
            data-cy={`access-token-preset-${token.id}-${preset.value}`}
          >
            {preset.label}
          </Button>
        ))}
      </ButtonGroup>

      <ButtonGroup size="xsmall" aria-label="Temporary access token preset">
        {temporaryDurations.map((duration) => (
          <Button
            key={duration.label}
            color="light"
            size="xsmall"
            className="!static !z-auto"
            disabled={!canTemporarilyElevate || mutation.isLoading}
            onClick={() =>
              mutation.mutate({
                tokenId: token.id,
                accessPreset: currentPreset,
                temporaryAccessPreset: 'manage',
                temporaryAccessExpiresAt: now + duration.seconds,
              })
            }
            title={`Grant Manage access for ${duration.label}`}
            data-cy={`access-token-temporary-manage-${token.id}-${duration.label}`}
          >
            Manage {duration.label}
          </Button>
        ))}
      </ButtonGroup>

      {hasTemporaryAccess && token.temporaryAccessExpiresAt && (
        <>
          <Badge type="warn" size="sm">
            Manage until{' '}
            {new Date(token.temporaryAccessExpiresAt * 1000).toLocaleString()}
          </Badge>
          <Button
            color="light"
            size="xsmall"
            disabled={mutation.isLoading}
            onClick={() =>
              mutation.mutate({
                tokenId: token.id,
                accessPreset: currentPreset,
                temporaryAccessPreset: '',
                temporaryAccessExpiresAt: 0,
              })
            }
            title="Revoke temporary access"
            data-cy={`access-token-temporary-revoke-${token.id}`}
          >
            Revoke
          </Button>
        </>
      )}
    </div>
  );
}
