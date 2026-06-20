import { describe, expect, it } from "vitest";
import type { AxiosResponse } from "axios";
import { ApiError, unwrap, unwrapList } from "./client";
import type { Envelope } from "./types";

function response<T>(data: Envelope<T>): AxiosResponse<Envelope<T>> {
  return {
    data,
    status: 200,
    statusText: "OK",
    headers: {},
    config: {} as never,
  };
}

describe("client helpers", () => {
  it("unwrap returns envelope data", async () => {
    await expect(unwrap(Promise.resolve(response({ success: true, data: { ok: true } })))).resolves.toEqual({
      ok: true,
    });
  });

  it("unwrap throws ApiError for failed envelopes", async () => {
    await expect(
      unwrap(
        Promise.resolve(
          response({
            success: false,
            data: null,
            error: { code: "VALIDATION_ERROR", message: "boom" },
          }),
        ),
      ),
    ).rejects.toEqual(new ApiError("VALIDATION_ERROR", "boom"));
  });

  it("unwrapList returns data and meta", async () => {
    await expect(
      unwrapList(
        Promise.resolve(
          response({
            success: true,
            data: [{ id: "1" }],
            meta: { total: 1, page: 1, limit: 20 },
          }),
        ),
      ),
    ).resolves.toEqual({
      data: [{ id: "1" }],
      meta: { total: 1, page: 1, limit: 20 },
    });
  });
});
