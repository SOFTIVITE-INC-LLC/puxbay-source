import { test, expect } from '@playwright/test';

test('has title', async ({ page }) => {
  await page.goto('/');
  // Basic check for the login page
  await expect(page).toHaveTitle(/Puxbay/i);
});

test('login flow', async ({ page }) => {
  await page.goto('/login');
  
  await page.fill('input[type="email"]', 'admin@softivite.com');
  await page.fill('input[type="password"]', 'password123');
  
  await page.click('button[type="submit"]');
  
  // Expect to be redirected to dashboard
  await expect(page).toHaveURL(/.*dashboard/);
});
