import { test, expect } from "@playwright/test";
import { mockBackend, seedAuth } from "./helpers/mockApi";

test.beforeEach(async ({ page }) => {
  await mockBackend(page);
  await seedAuth(page);
});

test("uploads a PDF: init -> presigned PUT -> complete", async ({ page }) => {
  await page.goto("/uploads/new");
  await expect(page.getByRole("heading", { name: /Upload PDF/i })).toBeVisible();

  // Provide an in-memory PDF file to the file input.
  await page.setInputFiles('input[type="file"]', {
    name: "de-thi.pdf",
    mimeType: "application/pdf",
    buffer: Buffer.from("%PDF-1.4 fake pdf bytes"),
  });

  await page.click('[data-testid="upload-submit"]');

  // After complete, the app navigates to the asset detail page.
  await expect(page).toHaveURL(/\/uploads\/asset-1$/);
  await expect(page.getByText("de-thi.pdf").first()).toBeVisible();
});
