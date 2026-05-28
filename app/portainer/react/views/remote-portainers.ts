import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { ListView } from '@/react/portainer/remote-portainers/ListView/ListView';
import { EditView } from '@/react/portainer/remote-portainers/EditView/EditView';
import { StacksView } from '@/react/portainer/remote-portainers/StacksView/StacksView';
import { StackEditView } from '@/react/portainer/remote-portainers/StackEditView/StackEditView';

export const remotePortainersModule = angular
  .module('portainer.app.react.views.remote-portainers', [])
  .component(
    'remotePortainersListView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ListView))), [])
  )
  .component(
    'remotePortainerEditView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EditView))), [])
  )
  .component(
    'remotePortainerStacksView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(StacksView))), [])
  )
  .component(
    'remotePortainerStackEditView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(StackEditView))), [])
  ).name;
