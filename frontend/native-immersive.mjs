import { chromium } from '@playwright/test';
import fs from 'node:fs';

const browser = await chromium.connectOverCDP('http://127.0.0.1:9222');
const ctx = browser.contexts()[0];
const page = ctx.pages().find(p => !p.url().startsWith('devtools')) ?? ctx.pages()[0];
await page.waitForTimeout(500);
await page.evaluate(() => {
  const b = document.querySelector('button[aria-label="Abrir modo inmersivo"]');
  b?.click();
});
await page.waitForTimeout(2500);
const state = await page.evaluate(() => {
  const d = document.querySelector('[role=dialog]');
  return { dialogs: document.querySelectorAll('[role=dialog]').length, text: (d?.innerText || '').slice(0, 200), bg: d ? getComputedStyle(d).backgroundColor : '' };
});
console.log(JSON.stringify(state, null, 2));
const client = await ctx.newCDPSession(page);
const shot = await client.send('Page.captureScreenshot', { format: 'png' });
fs.writeFileSync('C:/tmp/native-immersive.png', Buffer.from(shot.data, 'base64'));
await browser.close();
