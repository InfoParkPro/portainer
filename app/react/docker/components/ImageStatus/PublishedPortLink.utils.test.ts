import { describe, expect, it } from 'vitest';

import {
  buildPortURL,
  buildPublishedPortTargets,
  normalizeHost,
} from './PublishedPortLink.utils';

describe('normalizeHost', () => {
  it('strips scheme, path, query, and trailing slash', () => {
    expect(normalizeHost('https://example.com/portainer?x=1')).toBe(
      'example.com'
    );
  });

  it('keeps IPv6 hosts in URL-safe brackets', () => {
    expect(normalizeHost('fd7a:115c:a1e0::2f01:215a')).toBe(
      '[fd7a:115c:a1e0::2f01:215a]'
    );
  });
});

describe('buildPortURL', () => {
  it('builds explicit http and https URLs', () => {
    expect(buildPortURL('http', 'docker.local', 8088)).toBe(
      'http://docker.local:8088'
    );
    expect(buildPortURL('https', 'docker.local', 8088)).toBe(
      'https://docker.local:8088'
    );
  });
});

describe('buildPublishedPortTargets', () => {
  it('includes current host, public URL, and published host without duplicates', () => {
    const targets = buildPublishedPortTargets({
      currentHost: 'docker5.tail8add2.ts.net',
      publicURL: 'https://docker5.tail8add2.ts.net/portainer',
      publishedHost: '0.0.0.0',
      hostPort: 8088,
    });

    expect(targets.map((target) => target.host)).toEqual([
      'docker5.tail8add2.ts.net',
      '0.0.0.0',
    ]);
  });

  it('keeps a real published host when it differs from the current host', () => {
    const targets = buildPublishedPortTargets({
      currentHost: '192.168.0.1',
      publicURL: '',
      publishedHost: '100.103.33.90',
      hostPort: 8088,
    });

    expect(targets.map((target) => target.host)).toEqual([
      '192.168.0.1',
      '100.103.33.90',
    ]);
  });
});
