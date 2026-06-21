import { describe, expect, it } from "vitest";
import { tokenStore } from "./tokenStore";

describe("tokenStore", () => {
  it("stores access and refresh tokens in sessionStorage", () => {
    tokenStore.set("access-1", "refresh-1");
    expect(tokenStore.access).toBe("access-1");
    expect(tokenStore.refresh).toBe("refresh-1");
    expect(sessionStorage.getItem("edu_admin_access")).toBe("access-1");
    expect(sessionStorage.getItem("edu_admin_refresh")).toBe("refresh-1");
  });

  it("clears stored tokens", () => {
    tokenStore.set("access-1", "refresh-1");
    tokenStore.clear();
    expect(tokenStore.access).toBeNull();
    expect(tokenStore.refresh).toBeNull();
    expect(sessionStorage.getItem("edu_admin_access")).toBeNull();
    expect(sessionStorage.getItem("edu_admin_refresh")).toBeNull();
  });
});
