import { useEffect, useState } from 'react';
import { serverNow } from '../utils/serverClock';

/**
 * Re-renders on an interval so countdowns stay current. Returns server-corrected now (device time
 * plus the offset from the API's Date header), so a phone with a wrong clock still shows the right
 * expiry and "ends soon".
 */
export function useNow(intervalMs = 60_000): number {
  const [now, setNow] = useState(serverNow());
  useEffect(() => {
    setNow(serverNow());
    const id = setInterval(() => setNow(serverNow()), intervalMs);
    return () => clearInterval(id);
  }, [intervalMs]);
  return now;
}
