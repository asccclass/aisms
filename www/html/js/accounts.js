window.buildAccountQueryParams = function buildAccountQueryParams(status, q) {
  const params = new URLSearchParams();
  if (q) params.set('q', q);
  else if (status !== 'all') params.set('status', status);
  return params;
};

window.loadAccounts = async function loadAccounts() {
  const status = document.getElementById('status-filter').value;
  const q = document.getElementById('search-input').value.trim();
  let url = API + '/api/accounts';
  const params = buildAccountQueryParams(status, q);
  if ([...params].length) url += '?' + params;
  const r = await fetch(url);
  accounts = await r.json();
  renderAccounts(accounts);
};

window.renderAccounts = function renderAccounts(list) {
  const tbody = document.getElementById('accounts-body');
  if (!list || !list.length) {
    tbody.innerHTML = `<tr><td colspan="9"><div class="empty"><div class="icon">🔍</div><p>無符合資料</p></div></td></tr>`;
    return;
  }
  tbody.innerHTML = list.map(a => `
    <tr>
      <td style="color:var(--text-muted);font-size:12px;">#${a.id}</td>
      <td>${esc(a.system_name)}<br><span style="font-size:12px;color:var(--text-muted);">${esc(a.environment)}</span></td>
      <td>${esc(a.ip_address)}<br><span style="font-size:12px;color:var(--text-muted);">${esc(a.inventory_date)}</span></td>
      <td>${esc(a.department)}<br><span style="font-size:12px;color:var(--text-muted);">${esc(a.department_code)}</span></td>
      <td><strong>${esc(a.account_name)}</strong><br>${accountTypeBadge(a.account_type)}</td>
      <td>${esc(a.owner_name)}<br><span style="font-size:12px;color:var(--text-muted);">${esc(a.email)}</span></td>
      <td>${statusBadge(a.status)}</td>
      <td style="font-size:12px;color:var(--text-muted);">${esc(a.creator)}</td>
      <td><div class="actions"><button class="btn btn-ghost btn-sm" onclick="openEdit(${a.id})">✏️</button><button class="btn btn-warning btn-sm" onclick="openNotify(${a.id})" ${!a.email ? 'disabled' : ''}>📧</button><button class="btn btn-danger btn-sm" onclick="confirmDelete(${a.id}, '${escAttr(a.account_name)}')">🗑️</button></div></td>
    </tr>`).join('');
};

window.debounceSearch = function debounceSearch() {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(loadAccounts, 350);
};

window.getAccountFormSnapshot = function getAccountFormSnapshot() {
  const snapshot = {};
  ['system_name','environment','ip_address','inventory_date','account_name','account_type','department','department_code','owner_name','email','passphrase_rotate','status','remarks'].forEach(f => {
    snapshot[f] = document.getElementById('f-' + f)?.value || '';
  });
  return snapshot;
};

window.applyCurrentUserToAccountForm = function applyCurrentUserToAccountForm() {
  const profile = window.getCurrentUserProfile ? window.getCurrentUserProfile() : {};
  const setIfEmpty = (id, value) => {
    const el = document.getElementById(id);
    if (!el || el.value.trim() || !value) return;
    el.value = value;
  };

  setIfEmpty('f-owner_name', profile.name || '');
  setIfEmpty('f-department', profile.department || '');
  setIfEmpty('f-email', profile.email || '');
};

window.openCreate = async function openCreate() {
  const loaded = await ensureAccountModalLoaded();
  if (!loaded) return;
  editingId = null;
  document.getElementById('modal-title').textContent = '新增特殊權限帳號';
  clearForm();
  document.getElementById('f-inventory_date').value = new Date().toISOString().slice(0,10).replace(/-/g,'');
  applyCurrentUserToAccountForm();
  setModalSnapshotSource('account-modal', getAccountFormSnapshot);
  captureModalBaseline('account-modal');
  document.getElementById('account-modal').classList.add('open');
};

window.openEdit = async function openEdit(id) {
  const loaded = await ensureAccountModalLoaded();
  if (!loaded) return;
  editingId = id;
  document.getElementById('modal-title').textContent = '編輯帳號 #' + id;
  const r = await fetch(API + '/api/accounts/' + id);
  fillForm(await r.json());
  setModalSnapshotSource('account-modal', getAccountFormSnapshot);
  captureModalBaseline('account-modal');
  document.getElementById('account-modal').classList.add('open');
};

window.fillForm = function fillForm(a) {
  ['system_name','environment','ip_address','inventory_date','account_name','account_type','department','department_code','owner_name','email','passphrase_rotate','status','remarks'].forEach(f => {
    const el = document.getElementById('f-' + f);
    if (el) el.value = a[f] || '';
  });
};

window.clearForm = function clearForm() {
  ['system_name','ip_address','inventory_date','account_name','department','department_code','owner_name','email','remarks'].forEach(f => {
    const el = document.getElementById('f-' + f);
    if (el) el.value = '';
  });
  document.getElementById('f-account_type').value = '系統管理用';
  document.getElementById('f-environment').value = '正式區';
  document.getElementById('f-passphrase_rotate').value = '是';
  document.getElementById('f-status').value = 'active';
};

window.saveAccount = async function saveAccount() {
  const payload = {};
  ['system_name','environment','ip_address','inventory_date','account_name','account_type','department','department_code','owner_name','email','passphrase_rotate','status','remarks'].forEach(f => payload[f] = document.getElementById('f-' + f).value);
  if (!payload.system_name || !payload.account_name) return toast('請填寫必填欄位', 'error');
  const url = editingId ? API + '/api/accounts/' + editingId : API + '/api/accounts';
  const method = editingId ? 'PUT' : 'POST';
  const r = await fetch(url, { method, headers: {'Content-Type':'application/json'}, body: JSON.stringify(payload) });
  if (!r.ok) {
    const err = await r.json();
    return toast('錯誤：' + (err.error || '操作失敗'), 'error');
  }
  toast(editingId ? '帳號已更新 ✅' : '帳號已新增 ✅', 'success');
  captureModalBaseline('account-modal');
  closeModal('account-modal');
  await Promise.all([loadAccounts(), loadDashboardPage()]);
};

window.openNotify = async function openNotify(id) {
  const loaded = await ensureNotifyModalLoaded();
  if (!loaded) return;
  notifyId = id;
  const a = accounts.find(x => x.id === id);
  document.getElementById('notify-info-box').innerHTML = a ? `<div><strong>系統：</strong>${esc(a.system_name)} (${esc(a.ip_address)})</div><div><strong>帳號：</strong>${esc(a.account_name)} [${esc(a.account_type)}]</div><div><strong>使用者：</strong>${esc(a.owner_name)}</div><div><strong>Email：</strong>${esc(a.email)}</div>` : '載入中...';
  document.getElementById('notify-modal').classList.add('open');
};

window.doNotify = async function doNotify() {
  const r = await fetch(API + '/api/accounts/' + notifyId + '/notify', { method: 'POST' });
  const j = await r.json();
  closeModal('notify-modal');
  if (!r.ok) return toast('發送失敗：' + (j.error || '未知錯誤'), 'error');
  toast('通知已發送至 ' + j.email, 'success');
  loadAccounts();
};

window.notifyAll = async function notifyAll() {
  const btn = event.currentTarget;
  showConfirmDialog({
    title: '📧 批次通知確認',
    message: '將對所有「使用中」且有 Email 的帳號發送確認通知，確定繼續？',
    confirmLabel: '確認發送',
    confirmClass: 'btn btn-warning',
    onConfirm: async () => {
      btn.disabled = true;
      btn.textContent = '發送中…';
      const r = await fetch(API + '/api/notify-all', { method: 'POST' });
      const j = await r.json();
      btn.disabled = false;
      btn.innerHTML = '📧 批次通知所有使用中帳號';
      toast(`發送完成：成功 ${j.sent} 封，失敗 ${j.failed} 封`, j.failed ? 'error' : 'success');
      loadDashboardPage();
    }
  });
};

window.confirmDelete = function confirmDelete(id, name) {
  deleteId = id;
  showConfirmDialog({
    title: '⚠️ 確認刪除',
    message: `確定要刪除帳號「${name}」(#${id})？此操作無法復原。`,
    confirmLabel: '確定刪除',
    confirmClass: 'btn btn-danger',
    onConfirm: doDelete
  });
};

window.doDelete = async function doDelete() {
  const r = await fetch(API + '/api/accounts/' + deleteId, { method: 'DELETE' });
  closeConfirm();
  if (!r.ok) return toast('刪除失敗', 'error');
  toast('帳號已刪除', 'success');
  await Promise.all([loadAccounts(), loadDashboardPage()]);
};

window.loadLogs = async function loadLogs() {
  const r = await fetch(API + '/api/notification-logs');
  const logs = await r.json();
  const tbody = document.getElementById('logs-body');
  if (!logs.length) {
    tbody.innerHTML = `<tr><td colspan="7"><div class="empty"><div class="icon">📭</div><p>尚無通知記錄</p></div></td></tr>`;
    return;
  }
  tbody.innerHTML = logs.map(l => `<tr><td>${l.id}</td><td>${l.account_id}</td><td>${esc(l.account_name)}</td><td>${esc(l.email)}</td><td style="font-size:12px;">${new Date(l.sent_at).toLocaleString('zh-TW')}</td><td><span class="log-status-${l.status}">${l.status === 'sent' ? '✅ 成功' : '❌ 失敗'}</span></td><td style="font-size:12px;color:var(--text-muted);">${esc(l.message)}</td></tr>`).join('');
};
