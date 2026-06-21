import { describe, expect, it } from "vitest";
import { decodeJwt } from "./jwt";

function makeJwt(payload: object): string {
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const body = btoa(JSON.stringify(payload));
  return `${header}.${body}.signature`;
}

describe("decodeJwt", () => {
  it("parses a valid jwt payload", () => {
    const token = makeJwt({ sub: "user-1", role: "ADMIN", exp: 9999999999 });
    expect(decodeJwt(token)).toEqual({ sub: "user-1", role: "ADMIN", exp: 9999999999 });
  });

  it("returns null for malformed tokens", () => {
    expect(decodeJwt("not-a-jwt")).toBeNull();
    expect(decodeJwt("a.b.c")).toBeNull();
  });

  it("returns null for an empty string", () => {
    expect(decodeJwt("")).toBeNull();
  });

  it("decodes base64url payloads containing - and _", () => {
    // Build a payload whose base64 contains url-unsafe chars, then url-encode it.
    const payload = { sub: "u/+1", role: "ADMIN", exp: 9999999999 };
    const std = btoa(JSON.stringify(payload));
    const urlSafe = std.replace(/\+/g, "-").replace(/\//g, "_");
    const token = `${btoa("{}")}.${urlSafe}.sig`;
    expect(decodeJwt(token)).toEqual(payload);
  });
});
