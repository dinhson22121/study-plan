import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { AxiosAdapter } from "axios";
import axios from "axios";
import { api } from "./client";
import { tokenStore } from "@/auth/tokenStore";

/**
 * Exercises the 401 -> refresh -> retry interceptor on the real `api` instance.
 *
 * The retried request goes back through the `api` adapter, while the refresh
 * call uses a bare `axios.post`. We stub the adapter for retried requests and
 * stub `axios.post` for the refresh endpoint, so no network is touched.
 */
describe("api client 401 refresh flow", () => {
  let adapter: ReturnType<typeof vi.fn>;
  let originalAdapter: AxiosAdapter | undefined;

  beforeEach(() => {
    tokenStore.clear();
    originalAdapter = api.defaults.adapter as AxiosAdapter | undefined;
    adapter = vi.fn();
    api.defaults.adapter = adapter as unknown as AxiosAdapter;
  });

  afterEach(() => {
    api.defaults.adapter = originalAdapter;
    vi.restoreAllMocks();
    tokenStore.clear();
  });

  function unauthorized(config: unknown) {
    return Promise.reject({
      config,
      response: { status: 401, data: { success: false, error: { code: "UNAUTHORIZED", message: "expired" } } },
      isAxiosError: true,
    });
  }

  it("refreshes the token once and retries the original request", async () => {
    tokenStore.set("old-access", "refresh-1");

    // First call 401s, retried call succeeds.
    adapter
      .mockImplementationOnce((config) => unauthorized(config))
      .mockImplementationOnce((config) => Promise.resolve({
        data: { success: true, data: { ok: true } },
        status: 200,
        statusText: "OK",
        headers: {},
        config,
      }));

    const postSpy = vi.spyOn(axios, "post").mockResolvedValue({
      data: { success: true, data: { access_token: "new-access", refresh_token: "refresh-2" } },
    } as never);

    const res = await api.get("/admin/uploads");

    expect(res.data).toEqual({ success: true, data: { ok: true } });
    expect(postSpy).toHaveBeenCalledTimes(1);
    expect(tokenStore.access).toBe("new-access");
    expect(tokenStore.refresh).toBe("refresh-2");
    // retried request carries the fresh bearer token
    const retriedConfig = adapter.mock.calls[1][0];
    expect(retriedConfig.headers.Authorization).toBe("Bearer new-access");
  });

  it("clears tokens and triggers onUnauthorized when refresh fails", async () => {
    tokenStore.set("old-access", "refresh-1");
    adapter.mockImplementation((config) => unauthorized(config));
    vi.spyOn(axios, "post").mockRejectedValue(new Error("refresh boom"));

    await expect(api.get("/admin/uploads")).rejects.toMatchObject({ code: "UNAUTHORIZED" });
    expect(tokenStore.access).toBeNull();
    expect(tokenStore.refresh).toBeNull();
  });

  it("does not loop infinitely on repeated 401s (retry attempted once)", async () => {
    tokenStore.set("old-access", "refresh-1");
    // Always 401, even after a successful refresh.
    adapter.mockImplementation((config) => unauthorized(config));
    vi.spyOn(axios, "post").mockResolvedValue({
      data: { success: true, data: { access_token: "new-access", refresh_token: "refresh-2" } },
    } as never);

    await expect(api.get("/admin/uploads")).rejects.toBeTruthy();
    // original + one retry only
    expect(adapter).toHaveBeenCalledTimes(2);
  });

  it("shares a single refresh across concurrent 401s", async () => {
    tokenStore.set("old-access", "refresh-1");
    let callCount = 0;
    adapter.mockImplementation((config) => {
      callCount += 1;
      // first two calls (the two originals) 401; retries succeed.
      const retried = (config as { _retry?: boolean })._retry;
      if (retried) {
        return Promise.resolve({ data: { success: true, data: { ok: true } }, status: 200, statusText: "OK", headers: {}, config });
      }
      return unauthorized(config);
    });

    const postSpy = vi.spyOn(axios, "post").mockImplementation(
      () =>
        new Promise((resolve) =>
          setTimeout(
            () =>
              resolve({
                data: { success: true, data: { access_token: "new-access", refresh_token: "refresh-2" } },
              } as never),
            10,
          ),
        ),
    );

    await Promise.all([api.get("/a"), api.get("/b")]);

    // Single-flight: only one refresh POST despite two concurrent 401s.
    expect(postSpy).toHaveBeenCalledTimes(1);
    expect(callCount).toBe(4); // 2 originals + 2 retries
  });
});
