import {defineConfig} from '@playwright/test'

export default defineConfig({
  testDir: './tests/ui',
  timeout: 30000,
  expect: {timeout: 8000},
  fullyParallel: false,
  workers: 1,
  reporter: [['list'], ['html', {open: 'never', outputFolder: 'test-results/report'}]],
  outputDir: 'test-results/runs',
  use: {baseURL: 'http://127.0.0.1:4177', channel: 'chrome', headless: true, trace: 'retain-on-failure', screenshot: 'only-on-failure'},
  webServer: {command: 'npm run dev -- --host 127.0.0.1 --port 4177 --strictPort', url: 'http://127.0.0.1:4177', reuseExistingServer: false, timeout: 30000},
})
