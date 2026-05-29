import { test, expect } from "@playwright/test";

test.describe("Login Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
  });

  test("renders login form with email and password fields", async ({ page }) => {
    await expect(page.getByText("Sign In")).toBeVisible();
    await expect(page.locator("#email")).toBeVisible();
    await expect(page.locator("#password")).toBeVisible();
    await expect(
      page.getByRole("button", { name: /sign in/i })
    ).toBeVisible();
  });

  test("shows validation error for invalid email", async ({ page }) => {
    await page.locator("#email").fill("invalid-email");
    await page.locator("#password").fill("password123");
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page.getByText("Invalid email address")).toBeVisible();
  });

  test("shows validation error for short password", async ({ page }) => {
    await page.locator("#email").fill("test@example.com");
    await page.locator("#password").fill("short");
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(
      page.getByText("Password must be at least 8 characters")
    ).toBeVisible();
  });

  test("has link to register page", async ({ page }) => {
    const registerLink = page.getByRole("link", { name: /register/i });
    await expect(registerLink).toBeVisible();
    await expect(registerLink).toHaveAttribute("href", "/register");
  });
});

test.describe("Register Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/register");
  });

  test("renders registration form", async ({ page }) => {
    await expect(page.getByText("Create Account")).toBeVisible();
    await expect(page.locator("#reg-email")).toBeVisible();
    await expect(page.locator("#reg-password")).toBeVisible();
    await expect(page.locator("#reg-confirm")).toBeVisible();
    await expect(page.locator("#terms")).toBeVisible();
  });

  test("shows validation for password mismatch", async ({ page }) => {
    await page.locator("#reg-email").fill("test@example.com");
    await page.locator("#reg-password").fill("password123");
    await page.locator("#reg-confirm").fill("different456");
    await page.locator("#terms").check();
    await page.getByRole("button", { name: /create account/i }).click();

    await expect(page.getByText("Passwords do not match")).toBeVisible();
  });

  test("has link to login page", async ({ page }) => {
    const loginLink = page.getByRole("link", { name: /sign in/i });
    await expect(loginLink).toBeVisible();
    await expect(loginLink).toHaveAttribute("href", "/login");
  });
});
