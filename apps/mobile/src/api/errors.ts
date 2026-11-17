/** An HTTP error response from the API. Network failures and timeouts are plain Errors. */
export class HttpError extends Error {
  readonly status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = 'HttpError';
    this.status = status;
  }
}

/**
 * True only when the server says the stored user is gone (404) or its token is no good (401, 403).
 * Offline, timeouts and 5xx keep the stored user: a fan at a counter with no signal keeps their deals.
 */
export function shouldDiscardStoredUser(err: unknown): boolean {
  return err instanceof HttpError && (err.status === 401 || err.status === 403 || err.status === 404);
}

/** True when the request never got an HTTP answer (offline or timed out), so a retry is safe to try. */
export function isNetworkError(err: unknown): boolean {
  return !(err instanceof HttpError);
}
