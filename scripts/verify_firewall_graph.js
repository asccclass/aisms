const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1366, height: 768 } });
  await page.context().addCookies([
    { name: 'admin_email', value: 'andyliu%2540as.edu.tw', domain: 'localhost', path: '/' },
    { name: 'admin_name', value: 'Andy', domain: 'localhost', path: '/' }
  ]);
  await page.goto('http://localhost:8090/firewall-graph', { waitUntil: 'networkidle' });
  await page.waitForTimeout(2500);
  const result = await page.evaluate(() => {
    const canvas = document.getElementById('firewall-graph-canvas');
    const gl = canvas.getContext('webgl2') || canvas.getContext('webgl');
    const pixels = new Uint8Array(4 * 80 * 80);
    gl.readPixels(
      Math.max(0, Math.floor(canvas.width / 2) - 40),
      Math.max(0, Math.floor(canvas.height / 2) - 40),
      80,
      80,
      gl.RGBA,
      gl.UNSIGNED_BYTE,
      pixels
    );
    let nonBackground = 0;
    for (let i = 0; i < pixels.length; i += 4) {
      if (pixels[i] !== 7 || pixels[i + 1] !== 17 || pixels[i + 2] !== 31) nonBackground++;
    }
    return {
      nodes: document.getElementById('graph-node-count').textContent,
      edges: document.getElementById('graph-edge-count').textContent,
      rules: document.getElementById('graph-rule-count').textContent,
      canvasWidth: canvas.width,
      canvasHeight: canvas.height,
      nonBackground
    };
  });
  const selectedResult = await page.evaluate(() => {
    const firstNode = document.querySelector('.node-item');
    firstNode.click();
    const visibleNodes = window.__firewallGraphDebug.visibleNodeCount();
    const visibleEdges = window.__firewallGraphDebug.visibleEdgeCount();
    return {
      selectedName: document.querySelector('#graph-details h3')?.textContent || '',
      visibleNodes,
      visibleEdges
    };
  });
  await page.mouse.move(420, 360);
  await page.mouse.down();
  await page.mouse.move(520, 430);
  await page.mouse.up();
  const dragResult = await page.evaluate(() => ({
    visibleNodes: window.__firewallGraphDebug.visibleNodeCount(),
    visibleEdges: window.__firewallGraphDebug.visibleEdgeCount()
  }));
  console.log(JSON.stringify(selectedResult, null, 2));
  console.log(JSON.stringify(dragResult, null, 2));
  console.log(JSON.stringify(result, null, 2));
  await page.screenshot({ path: 'logs/firewall-graph-verification.png', fullPage: true });
  await browser.close();
  if (
    !Number(result.rules) ||
    result.nonBackground < 50 ||
    selectedResult.visibleNodes >= Number(result.nodes) ||
    dragResult.visibleNodes !== selectedResult.visibleNodes ||
    dragResult.visibleEdges !== selectedResult.visibleEdges
  ) {
    process.exit(1);
  }
})();
