import { test, expect } from '@playwright/test';

test.describe('Storefront critical journeys', () => {
  test('should load the home page and display products', async ({ page }) => {
    await page.goto('/');

    // Check title or main header
    await expect(page).toHaveTitle(/Puxbay/);

    // Wait for product grid to be visible
    const productGrid = page.locator('app-catalog');
    await expect(productGrid).toBeVisible({ timeout: 10000 });
  });

  test('should be able to search for a product', async ({ page }) => {
    await page.goto('/');

    // Click search icon (assuming there's a button with material-symbols-outlined containing 'search')
    const searchButton = page.locator('button:has-text("search")').first();
    await searchButton.click();

    // The search overlay should appear
    const searchInput = page.locator('input[placeholder*="Search"]');
    await expect(searchInput).toBeVisible();

    // Type a query
    await searchInput.fill('shirt');
    
    // Check that some results appear (or an empty state)
    // For now we just verify we can type into it.
    await expect(searchInput).toHaveValue('shirt');
  });
});
