/**
 * AccessToken represents an API key
 */
export interface AccessToken {
  id: number;

  userId: number;

  description: string;

  accessPreset: AccessTokenAccessPreset;

  effectiveAccessPreset?: AccessTokenAccessPreset;

  temporaryAccessPreset?: AccessTokenAccessPreset;

  /** Unix timestamp (UTC) when temporary access expires */
  temporaryAccessExpiresAt?: number;

  /** API key identifier (7 char prefix) */
  prefix: string;

  /** Unix timestamp (UTC) when the API key was created */
  dateCreated: number;

  /** Unix timestamp (UTC) when the API key was last used */
  lastUsed: number;

  /** Digest represents SHA256 hash of the raw API key */
  digest?: string;
}

export type AccessTokenAccessPreset =
  | 'disabled'
  | 'read_only'
  | 'power'
  | 'manage';
