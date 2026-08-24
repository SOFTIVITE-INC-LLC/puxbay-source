# Visual Regression Testing (VRT) Workflow

This document outlines the workflow for running and maintaining Visual Regression Tests in Puxbay.

## Overview
We use Playwright's built-in visual comparison capabilities (`toHaveScreenshot`) to catch unintended UI changes across different browsers and form factors.

## Writing Tests
1. Add a test in `e2e/vrt.spec.ts`.
2. Navigate to the page/component to test.
3. Call `expect(page).toHaveScreenshot('snapshot-name.png');`

Example:
```typescript
import { test, expect } from '@playwright/test';

test('dashboard visual regression', async ({ page }) => {
  await page.goto('/dashboard');
  await expect(page).toHaveScreenshot('dashboard-initial.png');
});
```

## Running VRT
To run tests and compare against existing baselines:
```bash
npx playwright test --grep @vrt
```

## Updating Baselines
When intentional UI changes are made, update the reference screenshots:
```bash
npx playwright test --update-snapshots
```

## Reviewing Diff
If a visual test fails in CI, Playwright generates an HTML report with a diff viewer showing the expected, actual, and difference images. Always review the diffs locally or via CI artifacts before updating baselines.
