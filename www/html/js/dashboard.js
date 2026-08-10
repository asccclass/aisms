window.loadDashboardConfig = async function loadDashboardConfig() {
  const [formsRes, providersRes] = await Promise.all([
    fetch(API + '/api/dashboard-forms'),
    fetch(API + '/api/dashboard-providers')
  ]);
  dashboardForms = (await formsRes.json()).map(normalizeDashboardForm);
  dashboardProviders = await providersRes.json();
  if (!dashboardSelectedFormKey || !dashboardForms.some(f => f.key === dashboardSelectedFormKey)) {
    const firstEnabled = dashboardForms.find(f => f.enabled);
    dashboardSelectedFormKey = firstEnabled ? firstEnabled.key : (dashboardForms[0]?.key || '');
  }
};

window.loadStats = async function loadStats() {
  const r = await fetch(API + '/api/stats');
  const s = await r.json();
  document.getElementById('stat-forms').textContent = dashboardForms.filter(f => f.enabled).length;
  document.getElementById('stat-active').textContent = s.active || 0;
  document.getElementById('stat-pending').textContent = s.pending || 0;
  document.getElementById('stat-closed').textContent = s.closed || 0;
};

window.loadDashboardAccounts = async function loadDashboardAccounts() {
  const r = await fetch(API + '/api/accounts?status=all');
  accounts = await r.json();
};

window.loadDashboardRecords = async function loadDashboardRecords() {
  const enabledForms = dashboardForms.filter(f => f.enabled);
  const results = await Promise.all(enabledForms.map(async form => {
    const r = await fetch(API + '/api/dashboard-records?form_key=' + encodeURIComponent(form.key));
    const data = r.ok ? await r.json() : [];
    return [form.key, data];
  }));
  dashboardRecordsByForm = Object.fromEntries(results);
};

window.loadDashboardPage = async function loadDashboardPage() {
  await loadDashboardConfig();
  await Promise.all([loadDashboardAccounts(), loadDashboardRecords()]);
  await loadStats();
  renderDashboard();
};

window.loadFormData = function loadFormData(form) {
  return dashboardRecordsByForm[form.key] || [];
};

window.renderDashboard = function renderDashboard() {
  renderFormsGrid();
  renderSelectedForm();
};

window.renderFormsGrid = function renderFormsGrid() {
  const enabledForms = dashboardForms.filter(form => form.enabled);
  const html = enabledForms.map(form => {
    const formData = loadFormData(form);
    const total = formData.length;
    const active = countByStatus(formData, 'active');
    const pending = countByStatus(formData, 'pending');
    const lastUpdated = getLastUpdatedRecord(formData);
    return `
      <div class="form-tile ${form.key === dashboardSelectedFormKey ? 'active' : ''}">
        <div class="form-tile-head">
          <div>
            <div class="form-code">${esc(form.code)}</div>
            <h4>${esc(form.name)}</h4>
            <p>${esc(form.description)}</p>
          </div>
          <div>${pending > 0 ? `<span class="badge badge-pending">${esc(form.status_needs_attention_text)}</span>` : `<span class="badge badge-active">${esc(form.status_normal_text)}</span>`}</div>
        </div>
        <div class="form-tile-metrics">
          <div class="form-tile-metric"><div class="label">紀錄數</div><div class="value">${total}</div></div>
          <div class="form-tile-metric"><div class="label">使用中</div><div class="value">${active}</div></div>
          <div class="form-tile-metric"><div class="label">待確認</div><div class="value">${pending}</div></div>
        </div>
        <div style="margin-top:14px;display:flex;justify-content:space-between;align-items:center;gap:12px;">
          <span class="hint">${lastUpdated ? `最近更新：${formatDate(lastUpdated.updated_at || lastUpdated.created_at)}` : '尚無資料'}</span>
          <button class="btn btn-ghost btn-sm" onclick="selectDashboardForm('${escAttr(form.key)}')">查看明細</button>
        </div>
      </div>`;
  }).join('');
  document.getElementById('dashboard-forms-grid').innerHTML = html ? `<div class="forms-grid">${html}</div>` : `<div class="empty"><div class="icon">🗂️</div><p>尚無啟用中的表單</p></div>`;
};

window.selectDashboardForm = function selectDashboardForm(formKey) {
  dashboardSelectedFormKey = formKey;
  renderDashboard();
};

window.renderSelectedForm = function renderSelectedForm() {
  const form = dashboardForms.find(x => x.key === dashboardSelectedFormKey && x.enabled) || dashboardForms.find(x => x.enabled) || dashboardForms[0];
  if (!form) {
    document.getElementById('dashboard-form-summary').innerHTML = '';
    document.getElementById('dashboard-status-list').innerHTML = '';
    document.getElementById('dashboard-table').innerHTML = `<div class="empty"><div class="icon">🗂️</div><p>尚無表單設定</p></div>`;
    return;
  }
  const list = loadFormData(form);
  renderDashboardSummary(form, list);
  renderDashboardFocus(form, list);
  renderDashboardRecentList(form, list);
};

window.renderDashboardRecentList = function renderDashboardRecentList(form, list) {
  const top10 = [...list].sort((a, b) => new Date(b.updated_at || b.created_at || 0) - new Date(a.updated_at || a.created_at || 0)).slice(0, 10);
  const tbody = top10.map(a => `
    <div class="dashboard-row">
      <span style="color:var(--text-muted);width:28px;">#${a.id}</span>
      <div class="dashboard-row-main">
        <div class="dashboard-row-title">${esc(a.primary_name)} ${statusBadge(a.status)}</div>
        <div class="dashboard-row-meta">${esc(a.secondary_name) || '未分類'}<br>使用者：${esc(a.owner_name) || '未填寫'} ｜ 盤點日：${esc(a.inventory_date) || '未填寫'} ｜ 更新：${formatDateTime(a.updated_at || a.created_at)}</div>
      </div>
    </div>`).join('');
  document.getElementById('dashboard-detail-title').textContent = `${form.short_code} ${form.detail_title}`;
  document.getElementById('dashboard-detail-desc').textContent = `顯示 ${form.short_code} ${form.name} 的最近更新資料。`;
  document.getElementById('dashboard-table').innerHTML = tbody || `<div class="empty"><div class="icon">📋</div><p>${esc(form.empty_text)}</p></div>`;
};

window.renderDashboardSummary = function renderDashboardSummary(form, list) {
  const total = list.length;
  const pending = countByStatus(list, 'pending');
  const expired = countByStatus(list, 'expired');
  const withEmail = list.filter(a => a.email).length;
  const lastUpdated = getLastUpdatedRecord(list);
  document.getElementById('dashboard-form-summary').innerHTML = `
    <div class="form-summary-card">
      <div class="form-summary-head">
        <div>
          <div class="form-code">${esc(form.code)}</div>
          <h4>${esc(form.name)}</h4>
          <p>${esc(form.description)}</p>
        </div>
        <div>${pending > 0 ? `<span class="badge badge-pending">${esc(form.status_needs_attention_text)}</span>` : `<span class="badge badge-active">${esc(form.status_normal_text)}</span>`}</div>
      </div>
      <div class="summary-metrics">
        <div class="metric-box"><div class="label">表單紀錄數</div><div class="value">${total}</div></div>
        <div class="metric-box"><div class="label">最近更新</div><div class="value" style="font-size:16px;">${lastUpdated ? formatDateTime(lastUpdated.updated_at || lastUpdated.created_at) : '尚無資料'}</div></div>
        <div class="metric-box"><div class="label">待確認 / 已過期</div><div class="value">${pending} / ${expired}</div></div>
        <div class="metric-box"><div class="label">具 Email 帳號</div><div class="value">${withEmail}</div></div>
      </div>
    </div>`;
};

window.renderDashboardFocus = function renderDashboardFocus(form, list) {
  const active = countByStatus(list, 'active');
  const pending = countByStatus(list, 'pending');
  const closed = countByStatus(list, 'closed');
  const lastUpdated = getLastUpdatedRecord(list);
  const focus = form.focus_items || DEFAULT_FORM_FOCUS_ITEMS;
  document.getElementById('dashboard-focus-title').textContent = `${form.short_code} 重點提醒`;
  document.getElementById('dashboard-focus-desc').textContent = `快速檢視 ${form.short_code} ${form.name} 目前最需要處理的項目。`;
  document.getElementById('dashboard-status-list').innerHTML = `
    <div class="status-list">
      <div class="status-item"><div><div class="title">${esc(focus.active_title)}</div><div class="meta">${esc(focus.active_meta)}</div></div><div class="count">${active}</div></div>
      <div class="status-item"><div><div class="title">${esc(focus.pending_title)}</div><div class="meta">${esc(focus.pending_meta)}</div></div><div class="count" style="color:var(--warning);">${pending}</div></div>
      <div class="status-item"><div><div class="title">${esc(focus.closed_title)}</div><div class="meta">${esc(focus.closed_meta)}</div></div><div class="count" style="color:var(--danger);">${closed}</div></div>
      <div class="status-item"><div><div class="title">${esc(focus.recent_title)}</div><div class="meta">${lastUpdated ? `${esc(lastUpdated.primary_name)} / ${esc(lastUpdated.secondary_name)}` : '尚無資料'}</div></div><div class="count" style="font-size:14px;">${lastUpdated ? formatDate(lastUpdated.updated_at || lastUpdated.created_at) : '—'}</div></div>
    </div>`;
};
