import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AuthProvider, useAuth } from "./AuthContext";
import { tokenStore } from "./tokenStore";
import * as clientModule from "@/api/client";

vi.mock("@/api/auth", () => ({
  login: vi.fn(),
  logout: vi.fn(),
}));

import * as authApi from "@/api/auth";

function makeJwt(payload: object): string {
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const body = btoa(JSON.stringify(payload));
  return `${header}.${body}.sig`;
}

function Probe() {
  const { user, isAuthenticated, isAdmin, login, logout } = useAuth();
  return (
    <div>
      <span data-testid="auth">{String(isAuthenticated)}</span>
      <span data-testid="admin">{String(isAdmin)}</span>
      <span data-testid="uid">{user?.id ?? ""}</span>
      <button onClick={() => login("a@b.com", "pw")}>login</button>
      <button onClick={() => logout()}>logout</button>
    </div>
  );
}

describe("AuthContext", () => {
  beforeEach(() => {
    tokenStore.clear();
    vi.clearAllMocks();
  });

  afterEach(() => {
    tokenStore.clear();
  });

  it("starts unauthenticated with no token", () => {
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );
    expect(screen.getByTestId("auth").textContent).toBe("false");
  });

  it("hydrates an admin user from a valid stored token", () => {
    const token = makeJwt({ sub: "user-9", role: "ADMIN", exp: Math.floor(Date.now() / 1000) + 3600 });
    tokenStore.set(token, "r");
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );
    expect(screen.getByTestId("auth").textContent).toBe("true");
    expect(screen.getByTestId("admin").textContent).toBe("true");
    expect(screen.getByTestId("uid").textContent).toBe("user-9");
  });

  it("ignores an expired stored token", () => {
    const token = makeJwt({ sub: "user-9", role: "ADMIN", exp: Math.floor(Date.now() / 1000) - 10 });
    tokenStore.set(token, "r");
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );
    expect(screen.getByTestId("auth").textContent).toBe("false");
  });

  it("login stores tokens and sets the user", async () => {
    const token = makeJwt({ sub: "user-1", role: "ADMIN", exp: Math.floor(Date.now() / 1000) + 3600 });
    vi.mocked(authApi.login).mockResolvedValue({
      access_token: token,
      refresh_token: "refresh-1",
      expires_at: "",
    });
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );
    await userEvent.click(screen.getByText("login"));
    await waitFor(() => expect(screen.getByTestId("auth").textContent).toBe("true"));
    expect(tokenStore.access).toBe(token);
    expect(tokenStore.refresh).toBe("refresh-1");
  });

  it("logout calls the api, clears tokens, and resets the user", async () => {
    const token = makeJwt({ sub: "user-1", role: "ADMIN", exp: Math.floor(Date.now() / 1000) + 3600 });
    tokenStore.set(token, "refresh-1");
    vi.mocked(authApi.logout).mockResolvedValue(undefined);
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );
    expect(screen.getByTestId("auth").textContent).toBe("true");
    await userEvent.click(screen.getByText("logout"));
    await waitFor(() => expect(screen.getByTestId("auth").textContent).toBe("false"));
    expect(authApi.logout).toHaveBeenCalledWith("refresh-1");
    expect(tokenStore.access).toBeNull();
  });

  it("clears the user when the api triggers onUnauthorized", async () => {
    const token = makeJwt({ sub: "user-1", role: "ADMIN", exp: Math.floor(Date.now() / 1000) + 3600 });
    tokenStore.set(token, "refresh-1");
    const setOnUnauthorizedSpy = vi.spyOn(clientModule, "setOnUnauthorized");
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );
    expect(screen.getByTestId("auth").textContent).toBe("true");
    // Grab the callback registered by AuthProvider and invoke it as the api would.
    const cb = setOnUnauthorizedSpy.mock.calls.at(-1)?.[0];
    expect(cb).toBeTypeOf("function");
    act(() => cb?.());
    await waitFor(() => expect(screen.getByTestId("auth").textContent).toBe("false"));
  });
});
