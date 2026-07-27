import clsx from 'clsx';
import { TaskState } from 'docker-types';

import { taskStatusBadge } from '@/docker/filters/utils';

import { multiple } from '@@/datatables/filter-types';
import { filterHOC } from '@@/datatables/Filter';

import { getTaskDisplayStatus, getTaskDisplayStatusBadge } from '../utils';

import { columnHelper } from './helper';

export const status = columnHelper.accessor(
  (item) => getTaskDisplayStatus(item),
  {
    header: 'Status',
    enableColumnFilter: true,
    filterFn: multiple,
    meta: {
      filter: filterHOC('Filter by state'),
      width: 100,
    },
    cell({ getValue }) {
      const value = getValue();

      const badge =
        getTaskDisplayStatusBadge(value) || taskStatusBadge(value as TaskState);

      return <span className={clsx('label', `label-${badge}`)}>{value}</span>;
    },
  }
);
