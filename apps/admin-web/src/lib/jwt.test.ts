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
});
