import { formatDateTime } from '@vben/utils';

type TimestampValue = null | number | string | undefined;

export function formatTimestamp(value: TimestampValue, fallback = '-') {
  if (value === null || value === undefined || value === '') {
    return fallback;
  }
  return formatDateTime(value);
}
