import type { AxiosAdapter } from 'axios';

import { api, ApiError } from '@/api/client';

// Drive the real interceptors/unwrap logic through a fake transport adapter.
function setAdapter(adapter: AxiosAdapter) {
  api.raw.defaults.adapter = adapter;
}

function reply(status: number, body: unknown): ReturnType<AxiosAdapter> {
  return Promise.resolve({
    data: body,
    status,
    statusText: '',
    headers: {},
    config: {} as never,
  });
}

describe('api client envelope handling', () => {
  it('unwraps the data payload on success', async () => {
    setAdapter((config) => reply(200, { success: true, data: { id: '7', name: 'X' } }));
    const out = await api.get<{ id: string; name: string }>('/things/7');
    expect(out).toEqual({ id: '7', name: 'X' });
  });

  it('throws ApiError with code/status on an error envelope', async () => {
    setAdapter(() => reply(422, { success: false, error: { code: 'VALIDATION_ERROR', message: 'sai dữ liệu' } }));
    await expect(api.post('/things', {})).rejects.toMatchObject({
      code: 'VALIDATION_ERROR',
      status: 422,
    });
    await expect(api.post('/things', {})).rejects.toBeInstanceOf(ApiError);
  });

  it('returns list data with pagination meta', async () => {
    setAdapter(() => reply(200, { success: true, data: [1, 2], meta: { total: 2, page: 1, limit: 20 } }));
    const { data, meta } = await api.getList<number[]>({ method: 'GET', url: '/items' });
    expect(data).toEqual([1, 2]);
    expect(meta?.total).toBe(2);
  });
});
