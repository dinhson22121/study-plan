import { test, expect } from "@playwright/test";
import { mockBackend } from "./helpers/mockApi";

test.beforeEach(async ({ page }) => {
  await mockBackend(page);
});

test("rejects invalid credentials", async ({ page }) => {
  await page.goto("/login");
  await page.fill("#email", "admin@example.com");
  await page.fill("#password", "wrongpass1");
  await page.click('[data-testid="login-form"] button[type="submit"]');
  await expect(page.getByText("Sai thông tin đăng nhập")).toBeVisible();
  await expect(page).toHaveURL(/\/login/);
});

test("logs in and lands on the dashboard", async ({ page }) => {
  await page.goto("/login");
  await page.fill("#email", "admin@example.com");
  await page.fill("#password", "password1");
  await page.click('[data-testid="login-form"] button[type="submit"]');
  await expect(page).toHaveURL(/\/dashboard/);
  await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
});

test("redirects unauthenticated users to login", async ({ page }) => {
  await page.goto("/uploads");
  await expect(page).toHaveURL(/\/login/);
});
