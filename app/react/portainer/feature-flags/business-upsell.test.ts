import { describe, expect, test, vi } from 'vitest';

import { isBusinessUpsellHidden } from './business-upsell';

describe('isBusinessUpsellHidden', () => {
  test('returns true when the local build flag is enabled', () => {
    vi.stubEnv('PORTAINER_HIDE_BUSINESS_UPSELL', 'true');

    expect(isBusinessUpsellHidden()).toBe(true);
  });

  test('returns false when the local build flag is disabled', () => {
    vi.stubEnv('PORTAINER_HIDE_BUSINESS_UPSELL', 'false');

    expect(isBusinessUpsellHidden()).toBe(false);
  });
});
