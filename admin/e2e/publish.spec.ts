import { test, expect } from "@playwright/test";
import { mockBackend, seedAuth } from "./helpers/mockApi";

test.beforeEach(async ({ page }) => {
  await mockBackend(page);
  await seedAuth(page);
});

test("publishes all drafts via the confirm dialog", async ({ page }) => {
  await page.goto("/uploads/asset-1/drafts");

  // Pick a publish target: subject -> chapter -> topic in the TopicSelector.
  await page.selectOption("select >> nth=0", "subj-1");
  await page.selectOption("select >> nth=1", "chap-1");
  await page.selectOption("select >> nth=2", "topic-1");

  await page.click('[data-testid="publish-all"]');

  // Confirm dialog appears with the publish prompt.
  await expect(page.getByText(/Publish toàn bộ draft/)).toBeVisible();
  await page.click('[data-testid="confirm-action"]');

  // Success toast confirms the publish completed.
  await expect(page.getByText(/Đã publish/)).toBeVisible();
});
