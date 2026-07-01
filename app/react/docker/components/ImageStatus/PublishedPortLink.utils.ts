export type PublishedPortScheme = 'http' | 'https';

export type PublishedPortTarget = {
  label: string;
  host: string;
  address: string;
  httpURL: string;
  httpsURL: string;
};

type BuildTargetsOptions = {
  currentHost?: string;
  publicURL?: string;
  publishedHost?: string;
  hostPort?: string | number;
};

export function buildPublishedPortTargets({
  currentHost,
  publicURL,
  publishedHost,
  hostPort,
}: BuildTargetsOptions) {
  const candidates = [
    { label: 'Current host', host: currentHost },
    { label: 'Environment URL', host: publicURL },
    { label: 'Published host', host: publishedHost },
  ];

  const seen = new Set<string>();
  const targets: PublishedPortTarget[] = [];

  candidates.forEach((candidate) => {
    const host = normalizeHost(candidate.host);
    const dedupeKey = host.toLowerCase();

    if (!host || seen.has(dedupeKey)) {
      return;
    }

    seen.add(dedupeKey);
    targets.push({
      label: candidate.label,
      host,
      address: buildPortAddress(host, hostPort),
      httpURL: buildPortURL('http', host, hostPort),
      httpsURL: buildPortURL('https', host, hostPort),
    });
  });

  return targets;
}

export function buildPortURL(
  scheme: PublishedPortScheme,
  host: string,
  hostPort?: string | number
) {
  return `${scheme}://${buildPortAddress(host, hostPort)}`;
}

export function buildPortAddress(host: string, hostPort?: string | number) {
  return `${host}:${hostPort}`;
}

export function normalizeHost(value?: string) {
  if (!value) {
    return '';
  }

  const host = stripURLParts(value.trim());

  if (isBracketedIPv6(host) || !isIPv6(host)) {
    return host;
  }

  return `[${host}]`;
}

function stripURLParts(value: string) {
  const withScheme = /^[a-z][a-z0-9+.-]*:\/\//i.test(value)
    ? value
    : `http://${value}`;

  try {
    return new URL(withScheme).hostname;
  } catch {
    return value.replace(/\/+$/, '');
  }
}

function isIPv6(host: string) {
  return host.includes(':');
}

function isBracketedIPv6(host: string) {
  return host.startsWith('[') && host.endsWith(']');
}
