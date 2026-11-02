const puppeteer = require('puppeteer');
const path = require('path');
const fs = require('fs');
const { execSync } = require('child_process');

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function capture() {
  const outputDir = path.join(__dirname, '..', 'docs', 'images');
  const framesDir = path.join(__dirname, '..', 'docs', 'images', 'frames');
  if (!fs.existsSync(outputDir)) fs.mkdirSync(outputDir, { recursive: true });
  if (!fs.existsSync(framesDir)) fs.mkdirSync(framesDir, { recursive: true });

  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 960, deviceScaleFactor: 2 });

  console.log('Navigating to MaskForge Studio at http://127.0.0.1:9888 ...');
  await page.goto('http://127.0.0.1:9888', { waitUntil: 'domcontentloaded' });
  await sleep(2000); // Allow initial schema scan to populate

  // Helper to click tab by text
  const clickTab = async (text) => {
    const buttons = await page.$$('button');
    for (const b of buttons) {
      const txt = await page.evaluate(el => el.textContent, b);
      if (txt && txt.includes(text)) {
        await b.click();
        await sleep(800);
        return true;
      }
    }
    return false;
  };

  let frameIdx = 1;
  const saveFrame = async () => {
    const pad = String(frameIdx++).padStart(2, '0');
    await page.screenshot({ path: path.join(framesDir, `frame_${pad}.png`) });
  };

  // 1. Schema & PII Matrix
  console.log('1/7 Capturing Schema & PII Matrix...');
  await page.screenshot({ path: path.join(outputDir, 'maskforge_schema_matrix.png') });
  await saveFrame();

  // 2. Rule Configurator
  console.log('2/7 Capturing Rule Configurator...');
  await clickTab('Rule Configurator');
  // Click mask button to trigger live preview
  const maskButtons = await page.$$('button');
  for (const b of maskButtons) {
    const txt = await page.evaluate(el => el.textContent, b);
    if (txt && txt.trim() === 'Mask') {
      await b.click();
      await sleep(600);
      break;
    }
  }
  await page.screenshot({ path: path.join(outputDir, 'maskforge_rule_editor.png') });
  await saveFrame();

  // 3. Before / After Diff
  console.log('3/7 Capturing Before / After Diff...');
  await clickTab('Before / After Diff');
  await page.screenshot({ path: path.join(outputDir, 'maskforge_diff_preview.png') });
  await saveFrame();

  // 4. Referential Integrity
  console.log('4/7 Capturing Referential Integrity...');
  await clickTab('Referential Integrity');
  // Click Assert Referential Join Match
  const refButtons = await page.$$('button');
  for (const b of refButtons) {
    const txt = await page.evaluate(el => el.textContent, b);
    if (txt && txt.includes('Assert Referential Join Match')) {
      await b.click();
      await sleep(800);
      break;
    }
  }
  await page.screenshot({ path: path.join(outputDir, 'maskforge_referential.png') });
  await saveFrame();

  // 5. Stream Simulator
  console.log('5/7 Capturing Stream Simulator...');
  await clickTab('Stream Simulator');
  // Click Run Streaming Pipe
  const streamButtons = await page.$$('button');
  for (const b of streamButtons) {
    const txt = await page.evaluate(el => el.textContent, b);
    if (txt && txt.includes('Run Streaming Pipe')) {
      await b.click();
      await sleep(1000);
      break;
    }
  }
  await page.screenshot({ path: path.join(outputDir, 'maskforge_stream_simulator.png') });
  await saveFrame();

  // 6. Zero-Leak Verifier
  console.log('6/7 Capturing Zero-Leak Verifier...');
  await clickTab('Zero-Leak Verifier');
  // Click Run Security Audit
  const auditButtons = await page.$$('button');
  for (const b of auditButtons) {
    const txt = await page.evaluate(el => el.textContent, b);
    if (txt && txt.includes('Run Security Audit')) {
      await b.click();
      await sleep(800);
      break;
    }
  }
  await page.screenshot({ path: path.join(outputDir, 'maskforge_leak_verifier.png') });
  await saveFrame();

  // Compile Animated Demo GIF using ffmpeg (desktop frames 1-6)
  console.log('Compiling maskforge_demo.gif via ffmpeg...');
  const gifPath = path.join(outputDir, 'maskforge_demo.gif');
  try {
    const cmd = `ffmpeg -y -framerate 0.65 -i "${path.join(framesDir, 'frame_%02d.png')}" -vf "scale=1280:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse=dither=bayer" "${gifPath}"`;
    execSync(cmd, { stdio: 'inherit' });
    console.log('✓ Successfully compiled maskforge_demo.gif!');
  } catch (e) {
    console.error('Failed to compile GIF with ffmpeg:', e.message);
  }

  // 7. Mobile Viewport (iPhone 14 / 390x844)
  console.log('7/7 Capturing Mobile Viewport...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2 });
  await clickTab('Schema & PII Matrix');
  await sleep(800);
  await page.screenshot({ path: path.join(outputDir, 'maskforge_mobile_responsive.png') });

  await browser.close();
  console.log('All screenshots and demo assets captured successfully!');

  // Clean up frames directory
  try {
    const files = fs.readdirSync(framesDir);
    for (const f of files) {
      fs.unlinkSync(path.join(framesDir, f));
    }
    fs.rmdirSync(framesDir);
  } catch {
    // ignore
  }
}

capture().catch((err) => {
  console.error('Screenshot capture failed:', err);
  process.exit(1);
});
