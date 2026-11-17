// Parsing saved copies (AsyncStorage) without ever letting a bad one block a fetch. Pure.

/**
 * Parse a saved copy. Returns null for a missing, corrupt or wrong-shaped cache, never throws, so
 * callers go on to fetch from the network whatever was saved.
 */
export function parseCache<T>(raw: string | null | undefined, valid: (v: any) => boolean): T | null {
  if (!raw) return null;
  try {
    const v = JSON.parse(raw);
    return v && typeof v === 'object' && valid(v) ? (v as T) : null;
  } catch {
    return null;
  }
}

/** Saved copy that belongs to one user: `{ userId, ... }`. */
export function parseUserCache<T extends { userId: string }>(
  raw: string | null | undefined,
  userId: string,
  valid: (v: any) => boolean
): T | null {
  return parseCache<T>(raw, (v) => v.userId === userId && valid(v));
}
