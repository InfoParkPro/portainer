import { ComponentType } from 'react';
import { Menu, MenuButton, MenuPopover } from '@reach/menu-button';
import { Copy, ExternalLink, Globe, LockKeyhole } from 'lucide-react';

import { Icon } from '@@/Icon';
import { useCopy } from '@@/buttons/CopyButton/useCopy';

import {
  buildPublishedPortTargets,
  PublishedPortTarget,
} from './PublishedPortLink.utils';

type Props = {
  hostURL?: string;
  publicURL?: string;
  hostPort?: string | number;
  containerPort?: string | number;
};

export function PublishedPortLink({
  hostURL,
  publicURL,
  hostPort,
  containerPort,
}: Props) {
  const currentHost =
    typeof window !== 'undefined' ? window.location.hostname : '';
  const targets = buildPublishedPortTargets({
    currentHost,
    publicURL,
    publishedHost: hostURL,
    hostPort,
  });

  if (targets.length === 0) {
    return (
      <span>
        {hostPort}:{containerPort}
      </span>
    );
  }

  return (
    <Menu>
      <MenuButton
        className="image-tag inline-flex items-center gap-1 border-0 bg-transparent p-0"
        data-cy="published-port-link"
        type="button"
      >
        {hostPort}:{containerPort}
        <Icon icon={ExternalLink} size="xs" />
      </MenuButton>
      <MenuPopover className="dropdown-menu min-w-[310px] p-0">
        <div className="bg-white py-1 th-highcontrast:bg-black th-dark:bg-black">
          {targets.map((target) => (
            <PublishedPortTargetRow key={target.host} target={target} />
          ))}
        </div>
      </MenuPopover>
    </Menu>
  );
}

function PublishedPortTargetRow({ target }: { target: PublishedPortTarget }) {
  const { handleCopy } = useCopy(target.address, 1000);

  return (
    <div className="flex items-center justify-between gap-3 border-b border-gray-5 px-3 py-2 last:border-b-0 th-highcontrast:border-gray-7 th-dark:border-gray-7">
      <div className="min-w-0">
        <div className="truncate text-sm font-medium">{target.host}</div>
        <div className="text-xs text-gray-7 th-dark:text-gray-5">
          {target.label}
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-1">
        <button
          type="button"
          className="btn btn-link btn-xs !ml-0 !px-1"
          title="Copy address"
          onClick={handleCopy}
          data-cy="published-port-copy-url"
        >
          <Icon icon={Copy} size="xs" />
        </button>
        <OpenURLButton
          url={target.httpURL}
          title="Open HTTP"
          icon={Globe}
          dataCy="published-port-open-http"
        />
        <OpenURLButton
          url={target.httpsURL}
          title="Open HTTPS"
          icon={LockKeyhole}
          dataCy="published-port-open-https"
        />
      </div>
    </div>
  );
}

function OpenURLButton({
  url,
  title,
  icon,
  dataCy,
}: {
  url: string;
  title: string;
  icon: ComponentType;
  dataCy: string;
}) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noreferrer"
      className="btn btn-link btn-xs !ml-0 !px-1"
      title={title}
      data-cy={dataCy}
    >
      <Icon icon={icon} size="xs" />
    </a>
  );
}
