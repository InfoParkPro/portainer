import { createColumnHelper } from '@tanstack/react-table';

import { isoDateFromTimestamp } from '@/portainer/filters/filters';

import { AccessToken } from '../../access-tokens/types';

import { AccessPresetSelector } from './AccessPresetSelector';

const columnHelper = createColumnHelper<AccessToken>();

export const columns = [
  columnHelper.accessor('description', {
    header: 'Description',
  }),
  columnHelper.accessor('prefix', {
    header: 'Prefix',
  }),
  columnHelper.accessor('accessPreset', {
    header: 'Access',
    cell: ({ row }) => <AccessPresetSelector token={row.original} />,
  }),
  columnHelper.accessor('dateCreated', {
    header: 'Created',
    cell: ({ getValue }) => isoDateFromTimestamp(getValue()),
  }),
  columnHelper.accessor('lastUsed', {
    header: 'Last Used',
    cell: ({ getValue }) => isoDateFromTimestamp(getValue()),
  }),
];
