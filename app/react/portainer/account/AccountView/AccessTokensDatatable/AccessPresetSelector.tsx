import clsx from 'clsx';

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

interface Props {
  token: AccessToken;
}

export function AccessPresetSelector({ token }: Props) {
  const mutation = useUpdateAccessTokenPresetMutation();
  const currentPreset = token.accessPreset || 'manage';

  return (
    <ButtonGroup size="xsmall" aria-label="Access token preset">
      {presets.map((preset) => (
        <Button
          key={preset.value}
          color="light"
          size="xsmall"
          className={clsx('!static !z-auto', {
            active: currentPreset === preset.value,
          })}
          disabled={
            currentPreset === preset.value || mutation.isLoading
          }
          onClick={() =>
            mutation.mutate({
              tokenId: token.id,
              accessPreset: preset.value,
            })
          }
          title={preset.title}
          data-cy={`access-token-preset-${token.id}-${preset.value}`}
        >
          {preset.label}
        </Button>
      ))}
    </ButtonGroup>
  );
}
