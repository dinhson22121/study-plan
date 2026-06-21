import { queryClient } from '@/lib/queryClient';
import { ApiError } from '@/api/client';

type RetryFn = (failureCount: number, error: Error) => boolean;

const retry = queryClient.getDefaultOptions().queries?.retry as RetryFn;

describe('queryClient retry policy', () => {
  it('never retries auth/permission/validation/not-found errors', () => {
    for (const status of [401, 403, 422, 404]) {
      expect(retry(0, new ApiError('X', 'x', status))).toBe(false);
    }
  });

  it('retries transient errors up to 2 times', () => {
    const err = new ApiError('INTERNAL', 'boom', 500);
    expect(retry(0, err)).toBe(true);
    expect(retry(1, err)).toBe(true);
    expect(retry(2, err)).toBe(false);
  });

  it('retries non-ApiError (network) errors', () => {
    expect(retry(0, new Error('network'))).toBe(true);
  });
});
