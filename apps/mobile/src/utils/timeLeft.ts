// Time-left labels for deal cards, free of native imports so bun tests can run them.

const MS_HOUR = 3_600_000;

/** "3h 28m left", "12m left", "1 day 7h left", "2 days left". */
export function formatTimeLeft(ms: number): string {
  if (ms <= 0) return 'Expired';
  const hours = Math.floor(ms / MS_HOUR);
  const minutes = Math.floor((ms % MS_HOUR) / 60_000);
  if (hours >= 48) return `${Math.floor(hours / 24)} days left`;
  if (hours >= 24) return `1 day ${hours - 24}h left`;
  if (hours > 0) return `${hours}h ${minutes}m left`;
  return `${Math.max(1, minutes)}m left`;
}

/** Spoken form: "3 hours 28 minutes left". */
export function spokenTimeLeft(ms: number): string {
  if (ms <= 0) return 'Expired';
  const hours = Math.floor(ms / MS_HOUR);
  const minutes = Math.floor((ms % MS_HOUR) / 60_000);
  const parts: string[] = [];
  if (hours >= 48) return `${Math.floor(hours / 24)} days left`;
  if (hours >= 24) return `1 day ${hours - 24} ${hours - 24 === 1 ? 'hour' : 'hours'} left`;
  if (hours > 0) parts.push(`${hours} ${hours === 1 ? 'hour' : 'hours'}`);
  parts.push(`${minutes} ${minutes === 1 ? 'minute' : 'minutes'}`);
  return `${parts.join(' ')} left`;
}
