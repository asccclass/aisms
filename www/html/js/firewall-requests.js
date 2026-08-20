(function () {
  const modalRootId = 'feature-firewall-request-modal-root';
  const modalPartialPath = '/partials/firewall-request-modal.html';
  const fields = [
    'legacy_form_number', 'system_name', 'action', 'purpose_type',
    'source_zone', 'source_zone2', 'source_ip', 'destination_zone',
    'destination_zone2', 'destination_ip', 'protocol_type', 'start_date',
    'end_date', 'request_date', 'rule_description', 'firewall_zone',
    'firewall_id', 'status', 'remarks'
  ];

  window.firewallRequests = [];
  window.firewallRequestEditingId = null;
  window.firewallRequestDeleteId = null;

  async function ensureFirewallRequestModalLoaded() {
    const root = document.getElementById(modalRootId);
    if (!root) return false;
    if (document.getElementById('firewall-request-modal')) return true;
    const response = await fetch(modalPartialPath);
    if (!response.ok) {
      toast('防火牆申請視窗載入失敗', 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  function todaySlash() {
    return new Date().toLocaleDateString('sv-SE').replace(/-/g, '/');
  }

  function getFirewallRequestFormSnapshot() {
    const snapshot = {};
    fields.forEach(field => {
      snapshot[field] = document.getElementById('fr-' + field)?.value || '';
    });
    return snapshot;
  }

  window.loadFirewallRequests = async function loadFirewallRequests() {
    const r = await fetch(API + '/api/firewall-requests');
    firewallRequests = await r.json();
    renderFirewallRequests(firewallRequests);
  };

  window.renderFirewallRequests = function renderFirewallRequests(list) {
    const tbody = document.getElementById('firewall-requests-body');
    if (!tbody) return;
    if (!list || !list.length) {
      tbody.innerHTML = `<tr><td colspan="10"><div class="empty"><div class="icon">🔥</div><p>尚無防火牆申請資料</p></div></td></tr>`;
      return;
    }
    tbody.innerHTML = list.map(item => `
      <tr>
        <td style="color:var(--text-muted);font-size:12px;">#${item.id}</td>
        <td>${esc(item.system_name)}<br><span class="hint">${esc(item.legacy_form_number || '—')}</span></td>
        <td>${esc(item.action)}<br><span class="hint">${esc(item.purpose_type)}</span></td>
        <td>${esc(item.source_zone)}${item.source_zone2 ? ' / ' + esc(item.source_zone2) : ''}<br><span class="hint mono">${esc(item.source_ip)}</span></td>
        <td>${esc(item.destination_zone)}${item.destination_zone2 ? ' / ' + esc(item.destination_zone2) : ''}<br><span class="hint mono">${esc(item.destination_ip)}</span></td>
        <td>${esc(item.protocol_type)}</td>
        <td>${esc(item.start_date)}<br><span class="hint">${esc(item.end_date)}</span></td>
        <td>${esc(item.firewall_zone)}<br><span class="hint mono">${esc(item.firewall_id)}</span></td>
        <td>${statusBadge(item.status)}</td>
        <td><div class="actions"><button class="btn btn-ghost btn-sm" onclick="openFirewallRequestEdit(${item.id})">✏️</button><button class="btn btn-danger btn-sm" onclick="confirmDeleteFirewallRequest(${item.id}, '${escAttr(item.system_name)}')">🗑️</button></div></td>
      </tr>`).join('');
  };

  window.openFirewallRequestCreate = async function openFirewallRequestCreate() {
    const loaded = await ensureFirewallRequestModalLoaded();
    if (!loaded) return;
    firewallRequestEditingId = null;
    document.getElementById('firewall-request-modal-title').textContent = '新增防火牆申請';
    clearFirewallRequestForm();
    document.getElementById('fr-request_date').value = todaySlash();
    setModalSnapshotSource('firewall-request-modal', getFirewallRequestFormSnapshot);
    captureModalBaseline('firewall-request-modal');
    document.getElementById('firewall-request-modal').classList.add('open');
  };

  window.openFirewallRequestEdit = async function openFirewallRequestEdit(id) {
    const loaded = await ensureFirewallRequestModalLoaded();
    if (!loaded) return;
    firewallRequestEditingId = id;
    document.getElementById('firewall-request-modal-title').textContent = '編輯防火牆申請 #' + id;
    const r = await fetch(API + '/api/firewall-requests/' + id);
    const data = await r.json();
    fillFirewallRequestForm(data);
    setModalSnapshotSource('firewall-request-modal', getFirewallRequestFormSnapshot);
    captureModalBaseline('firewall-request-modal');
    document.getElementById('firewall-request-modal').classList.add('open');
  };

  window.clearFirewallRequestForm = function clearFirewallRequestForm() {
    fields.forEach(field => {
      const el = document.getElementById('fr-' + field);
      if (el) el.value = '';
    });
    const display = document.getElementById('fr-id_display');
    if (display) display.value = '';
    document.getElementById('fr-status').value = 'active';
  };

  window.fillFirewallRequestForm = function fillFirewallRequestForm(item) {
    fields.forEach(field => {
      const el = document.getElementById('fr-' + field);
      if (el) el.value = item[field] || '';
    });
    const display = document.getElementById('fr-id_display');
    if (display) display.value = item.id || '';
  };

  window.saveFirewallRequest = async function saveFirewallRequest() {
    const payload = {};
    fields.forEach(field => {
      const el = document.getElementById('fr-' + field);
      payload[field] = el ? el.value.trim() : '';
    });
    if (!payload.system_name || !payload.action || !payload.source_ip || !payload.destination_ip) {
      return toast('請填寫系統名稱、執行動作、來源 IP 與目的地 IP', 'error');
    }
    const url = firewallRequestEditingId ? API + '/api/firewall-requests/' + firewallRequestEditingId : API + '/api/firewall-requests';
    const method = firewallRequestEditingId ? 'PUT' : 'POST';
    const r = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await r.json();
    if (!r.ok) return toast('儲存失敗：' + (data.error || '未知錯誤'), 'error');
    captureModalBaseline('firewall-request-modal');
    closeModal('firewall-request-modal');
    toast(firewallRequestEditingId ? '防火牆申請已更新' : '防火牆申請已新增', 'success');
    await Promise.all([loadFirewallRequests(), loadDashboardPage()]);
  };

  window.confirmDeleteFirewallRequest = function confirmDeleteFirewallRequest(id, name) {
    firewallRequestDeleteId = id;
    showConfirmDialog({
      title: '⚠️ 確認刪除',
      message: `確定要刪除防火牆申請「${name}」(#${id})？此操作無法復原。`,
      confirmLabel: '確定刪除',
      confirmClass: 'btn btn-danger',
      onConfirm: doDeleteFirewallRequest
    });
  };

  window.doDeleteFirewallRequest = async function doDeleteFirewallRequest() {
    const r = await fetch(API + '/api/firewall-requests/' + firewallRequestDeleteId, { method: 'DELETE' });
    closeConfirm();
    if (!r.ok) return toast('刪除防火牆申請失敗', 'error');
    toast('防火牆申請已刪除', 'success');
    await Promise.all([loadFirewallRequests(), loadDashboardPage()]);
  };
})();
