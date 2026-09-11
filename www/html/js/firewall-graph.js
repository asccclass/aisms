import * as THREE from 'https://cdn.jsdelivr.net/npm/three@0.168.0/build/three.module.js';

const esc = window.esc || (value => String(value ?? ''));
const escAttr = window.escAttr || esc;
const canvas = document.getElementById('firewall-graph-canvas');
const details = document.getElementById('graph-details');
const graph = {
  records: [],
  nodes: [],
  edges: [],
  nodeMap: new Map(),
  edgeMap: new Map(),
  scene: null,
  camera: null,
  renderer: null,
  group: null,
  raycaster: new THREE.Raycaster(),
  pointer: new THREE.Vector2(),
  selected: null,
  rotation: { x: -0.24, y: 0.42 },
  distance: 560,
  dragging: false,
  lastPointer: { x: 0, y: 0 }
};

window.loadFirewallGraph = async function loadFirewallGraph() {
  try {
    const response = await fetch(API + '/api/firewall-requests');
    if (!response.ok) throw new Error('load failed');
    graph.records = await response.json();
    buildGraph(graph.records || []);
    initScene();
    renderDetails();
  } catch (_) {
    details.innerHTML = '<h3>資料載入失敗</h3><p class="hint">請確認已登入且 04-042 API 可正常存取。</p>';
  }
};

window.resetFirewallGraphView = function resetFirewallGraphView() {
  graph.rotation = { x: -0.24, y: 0.42 };
  graph.distance = 560;
  graph.selected = null;
  updateCamera();
  highlightSelection();
  renderDetails();
};

window.filterFirewallGraph = function filterFirewallGraph(keyword) {
  const q = String(keyword || '').trim().toLowerCase();
  graph.nodes.forEach(node => {
    const matched = !q || node.searchText.includes(q);
    node.mesh.material.opacity = matched ? 1 : 0.18;
    node.label.material.opacity = matched ? 0.92 : 0.16;
  });
  graph.edges.forEach(edge => {
    const matched = !q || edge.searchText.includes(q);
    edge.line.material.opacity = matched ? 0.55 : 0.08;
  });
};

function buildGraph(records) {
  graph.nodes = [];
  graph.edges = [];
  graph.nodeMap = new Map();
  graph.edgeMap = new Map();

  records.forEach(record => {
    const systemName = normalizeName(record.system_name, '未命名主機');
    const sourceName = normalizeName(record.source_ip || record.source_zone, '未知來源');
    const destinationName = normalizeName(record.destination_ip || record.destination_zone, '未知目的');
    const system = getNode(systemName, 'host');
    const source = getNode(sourceName, 'source');
    const destination = getNode(destinationName, 'destination');

    system.rules.push(record);
    source.rules.push(record);
    destination.rules.push(record);
    connect(source, system, record);
    connect(system, destination, record);
  });

  layoutNodes();
  document.getElementById('graph-node-count').textContent = graph.nodes.length;
  document.getElementById('graph-edge-count').textContent = graph.edges.length;
  document.getElementById('graph-rule-count').textContent = records.length;
}

function normalizeName(value, fallback) {
  const text = String(value || '').trim();
  return text || fallback;
}

function getNode(name, type) {
  const key = `${type}:${name}`;
  if (graph.nodeMap.has(key)) return graph.nodeMap.get(key);
  const node = {
    id: key,
    name,
    type,
    rules: [],
    position: new THREE.Vector3(),
    mesh: null,
    label: null,
    searchText: `${name} ${type}`.toLowerCase()
  };
  graph.nodeMap.set(key, node);
  graph.nodes.push(node);
  return node;
}

function connect(from, to, record) {
  const key = `${from.id}->${to.id}`;
  let edge = graph.edgeMap.get(key);
  if (!edge) {
    edge = { from, to, records: [], line: null, searchText: '' };
    graph.edgeMap.set(key, edge);
    graph.edges.push(edge);
  }
  edge.records.push(record);
  edge.searchText = `${from.name} ${to.name} ${record.protocol_type || ''} ${record.firewall_id || ''}`.toLowerCase();
}

function layoutNodes() {
  const hosts = graph.nodes.filter(node => node.type === 'host');
  const sources = graph.nodes.filter(node => node.type === 'source');
  const destinations = graph.nodes.filter(node => node.type === 'destination');
  placeRing(hosts, 0, 0, 0, Math.max(80, hosts.length * 18), 0);
  placeRing(sources, -210, 0, -30, Math.max(90, sources.length * 12), 0.35);
  placeRing(destinations, 210, 0, 30, Math.max(90, destinations.length * 12), -0.2);
}

function placeRing(nodes, centerX, centerY, centerZ, radius, phase) {
  nodes.forEach((node, index) => {
    const angle = phase + (Math.PI * 2 * index) / Math.max(nodes.length, 1);
    const zWave = Math.sin(angle * 2.3) * 44;
    node.position.set(centerX + Math.cos(angle) * radius, centerY + Math.sin(angle) * radius, centerZ + zWave);
  });
}

function initScene() {
  if (graph.renderer) {
    graph.renderer.dispose();
    graph.group.clear();
  }
  graph.scene = new THREE.Scene();
  graph.scene.background = new THREE.Color(0x07111f);
  graph.group = new THREE.Group();
  graph.scene.add(graph.group);

  graph.camera = new THREE.PerspectiveCamera(48, 1, 1, 5000);
  graph.renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: false });
  graph.renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));

  const ambient = new THREE.AmbientLight(0xffffff, 0.72);
  const point = new THREE.PointLight(0x7dd3fc, 1.8, 1100);
  point.position.set(180, 260, 300);
  graph.scene.add(ambient, point);

  addGrid();
  drawEdges();
  drawNodes();
  bindEvents();
  resize();
  animate();
}

function addGrid() {
  const grid = new THREE.GridHelper(760, 24, 0x2d6ca2, 0x16324c);
  grid.rotation.x = Math.PI / 2;
  grid.material.opacity = 0.22;
  grid.material.transparent = true;
  graph.group.add(grid);
}

function drawNodes() {
  const geometry = new THREE.SphereGeometry(1, 32, 20);
  graph.nodes.forEach(node => {
    const color = node.type === 'host' ? 0x38bdf8 : node.type === 'source' ? 0x22c55e : 0xf59e0b;
    const scale = node.type === 'host' ? 13 : 8;
    const material = new THREE.MeshStandardMaterial({
      color,
      emissive: color,
      emissiveIntensity: node.type === 'host' ? 0.34 : 0.2,
      roughness: 0.42,
      transparent: true
    });
    node.mesh = new THREE.Mesh(geometry, material);
    node.mesh.position.copy(node.position);
    node.mesh.scale.setScalar(scale + Math.min(node.rules.length, 8) * 0.9);
    node.mesh.userData.node = node;
    node.label = createLabel(node.name, color);
    node.label.position.copy(node.position).add(new THREE.Vector3(0, scale + 10, 0));
    graph.group.add(node.mesh, node.label);
  });
}

function drawEdges() {
  graph.edges.forEach(edge => {
    const material = new THREE.LineBasicMaterial({
      color: 0x9cc9ff,
      transparent: true,
      opacity: 0.48
    });
    const points = [edge.from.position, edge.to.position];
    edge.line = new THREE.Line(new THREE.BufferGeometry().setFromPoints(points), material);
    graph.group.add(edge.line);
  });
}

function createLabel(text, color) {
  const canvasLabel = document.createElement('canvas');
  const context = canvasLabel.getContext('2d');
  const fontSize = 24;
  const label = text.length > 24 ? text.slice(0, 23) + '...' : text;
  context.font = `600 ${fontSize}px "Segoe UI", Arial`;
  const width = Math.ceil(context.measureText(label).width + 28);
  canvasLabel.width = width;
  canvasLabel.height = 44;
  context.font = `600 ${fontSize}px "Segoe UI", Arial`;
  context.fillStyle = 'rgba(7,17,31,0.72)';
  context.fillRect(0, 0, width, 44);
  context.fillStyle = `#${color.toString(16).padStart(6, '0')}`;
  context.fillText(label, 14, 29);
  const texture = new THREE.CanvasTexture(canvasLabel);
  const material = new THREE.SpriteMaterial({ map: texture, transparent: true, opacity: 0.92 });
  const sprite = new THREE.Sprite(material);
  sprite.scale.set(width * 0.32, 14, 1);
  return sprite;
}

function bindEvents() {
  window.addEventListener('resize', resize);
  canvas.onpointerdown = event => {
    graph.dragging = true;
    graph.lastPointer = { x: event.clientX, y: event.clientY };
    canvas.setPointerCapture(event.pointerId);
  };
  canvas.onpointermove = event => {
    if (!graph.dragging) return;
    const dx = event.clientX - graph.lastPointer.x;
    const dy = event.clientY - graph.lastPointer.y;
    graph.rotation.y += dx * 0.006;
    graph.rotation.x += dy * 0.006;
    graph.rotation.x = Math.max(-1.25, Math.min(1.25, graph.rotation.x));
    graph.lastPointer = { x: event.clientX, y: event.clientY };
    updateCamera();
  };
  canvas.onpointerup = event => {
    graph.dragging = false;
    canvas.releasePointerCapture(event.pointerId);
  };
  canvas.onclick = event => {
    const rect = canvas.getBoundingClientRect();
    graph.pointer.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    graph.pointer.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
    graph.raycaster.setFromCamera(graph.pointer, graph.camera);
    const hit = graph.raycaster.intersectObjects(graph.nodes.map(node => node.mesh))[0];
    graph.selected = hit ? hit.object.userData.node : null;
    highlightSelection();
    renderDetails();
  };
  canvas.onwheel = event => {
    event.preventDefault();
    graph.distance = Math.max(180, Math.min(1200, graph.distance + event.deltaY * 0.35));
    updateCamera();
  };
}

function resize() {
  const rect = canvas.parentElement.getBoundingClientRect();
  graph.renderer.setSize(rect.width, rect.height, false);
  graph.camera.aspect = rect.width / Math.max(rect.height, 1);
  graph.camera.updateProjectionMatrix();
  updateCamera();
}

function updateCamera() {
  const x = Math.sin(graph.rotation.y) * Math.cos(graph.rotation.x) * graph.distance;
  const y = Math.sin(graph.rotation.x) * graph.distance;
  const z = Math.cos(graph.rotation.y) * Math.cos(graph.rotation.x) * graph.distance;
  graph.camera.position.set(x, y, z);
  graph.camera.lookAt(0, 0, 0);
}

function highlightSelection() {
  graph.nodes.forEach(node => {
    const selected = graph.selected && node.id === graph.selected.id;
    node.mesh.material.emissiveIntensity = selected ? 0.85 : (node.type === 'host' ? 0.34 : 0.2);
  });
  graph.edges.forEach(edge => {
    const connected = graph.selected && (edge.from.id === graph.selected.id || edge.to.id === graph.selected.id);
    edge.line.material.opacity = graph.selected ? (connected ? 0.9 : 0.1) : 0.48;
  });
}

function renderDetails() {
  if (!graph.records.length) {
    details.innerHTML = '<h3>尚無防火牆資料</h3><p class="hint">04-042 匯入或建立資料後，這裡會呈現關聯圖。</p>';
    return;
  }
  if (!graph.selected) {
    details.innerHTML = `<h3>防火牆關聯概覽</h3><p class="hint">拖曳旋轉、滾輪縮放，點選節點查看關聯規則。</p>
      <div class="rule-list">${graph.nodes.slice(0, 10).map(node => `<div class="rule-item"><strong>${esc(node.name)}</strong><br><span class="hint">${node.rules.length} 筆關聯規則</span></div>`).join('')}</div>`;
    return;
  }
  const rules = graph.selected.rules.slice(0, 20);
  details.innerHTML = `<h3>${esc(graph.selected.name)}</h3><p class="hint">${graph.selected.rules.length} 筆關聯規則</p>
    <div class="rule-list">${rules.map(rule => `<div class="rule-item">
      <strong>${esc(rule.system_name || '未命名主機')}</strong><br>
      <span class="hint">${esc(rule.source_ip)} → ${esc(rule.destination_ip)}</span><br>
      ${esc(rule.protocol_type || '')}<br>
      <span class="hint">${esc(rule.firewall_id || '')}</span>
    </div>`).join('')}</div>`;
}

function animate() {
  requestAnimationFrame(animate);
  if (graph.group && !graph.dragging) graph.group.rotation.z += 0.0008;
  graph.renderer.render(graph.scene, graph.camera);
}
