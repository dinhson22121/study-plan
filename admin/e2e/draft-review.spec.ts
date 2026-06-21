import { test, expect } from "@playwright/test";
import { mockBackend, seedAuth } from "./helpers/mockApi";

test.beforeEach(async ({ page }) => {
  await mockBackend(page);
  await seedAuth(page);
});

test("shows draft questions for review", async ({ page }) => {
  await page.goto("/uploads/asset-1/drafts");
  await expect(page.getByRole("heading", { name: /Review draft questions/i })).toBeVisible();
  await expect(page.getByText("2 + 2 = ?")).toBeVisible();
  await expect(page.getByText(/Câu 1/)).toBeVisible();
});

test("navigates from asset detail into review & publish", async ({ page }) => {
  await page.goto("/uploads/asset-1");
  await expect(page.getByRole("heading", { name: "de-thi.pdf" })).toBeVisible();
  await page.click('[data-testid="review-publish"]');
  await expect(page).toHaveURL(/\/uploads\/asset-1\/drafts$/);
});
