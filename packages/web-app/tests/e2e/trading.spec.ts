import { test, expect } from "@playwright/test";

test.describe("Trading Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/trade");
  });

  test("renders the trading chart area", async ({ page }) => {
    // Chart container should be present
    const chartContainer = page.locator('[class*="col-span"]').first();
    await expect(chartContainer).toBeVisible();
  });

  test("order book displays bid and ask levels", async ({ page }) => {
    // Order book heading should be present
    await expect(page.getByText("Order Book")).toBeVisible();

    // Should have price values in the orderbook
    const prices = page.locator(".font-mono");
    expect(await prices.count()).toBeGreaterThan(0);
  });

  test("order form is interactive with tabs", async ({ page }) => {
    // Tab buttons should be visible
    await expect(page.getByRole("button", { name: "Buy" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Sell" })).toBeVisible();

    // Click Market tab
    await page.getByText("market", { exact: false }).first().click();

    // Quantity input should be present
    const quantityInput = page.locator('input[placeholder="0.0000"]');
    await expect(quantityInput).toBeVisible();
  });

  test("can switch between positions and orders tabs", async ({ page }) => {
    // Click Open Orders tab
    await page.getByText("Open Orders").click();

    // Should show orders table headers
    await expect(page.getByText("Pair").first()).toBeVisible();
  });

  test("time interval selector buttons are present", async ({ page }) => {
    await expect(page.getByRole("button", { name: "1h" })).toBeVisible();
    await expect(page.getByRole("button", { name: "1D" })).toBeVisible();
  });
});

test.describe("Responsive Layout", () => {
  test("shows bottom navigation on mobile", async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto("/trade");

    // Mobile nav should be visible on small screens
    const mobileNav = page.locator("nav.fixed.bottom-0");
    await expect(mobileNav).toBeVisible();
  });

  test("shows sidebar on desktop", async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto("/trade");

    // Sidebar should be visible on larger screens
    const sidebar = page.locator("aside");
    await expect(sidebar).toBeVisible();
  });
});
