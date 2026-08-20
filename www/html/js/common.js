window.API = '';
window.DEFAULT_FORM_FOCUS_ITEMS = {
  active_title: '有效資料',
  active_meta: '顯示此表單目前有效或啟用中的資料筆數',
  pending_title: '待處理事項',
  pending_meta: '顯示此表單待辦、待確認或待補件數量',
  closed_title: '已完成事項',
  closed_meta: '顯示此表單已完成、已結案或已停用數量',
  recent_title: '最近更新紀錄'
};

window.accounts = [];
window.dashboardForms = [];
window.dashboardProviders = [];
window.dashboardRecordsByForm = {};
window.dashboardSelectedFormKey = '';
window.editingId = null;
window.formEditingId = null;
window.notifyId = null;
window.deleteId = null;
window.searchTimer = null;
window.profileTargets = ['user-profile-dash', 'user-profile-acc', 'user-profile-forms', 'user-profile-platform', 'user-profile-firewall'];
window.profileExpanded = false;
window.currentUserProfile = null;
window.modalSnapshotGetters = {};
window.modalSnapshotBaselines = {};
window.confirmState = null;

window.normalizeDashboardForm = function normalizeDashboardForm(form) {
  return {
    detail_title: '最近資料清單',
    empty_text: '此表單尚未接入資料來源',
    status_normal_text: '狀況正常',
    status_needs_attention_text: '需追蹤',
    enabled: true,
    focus_items: { ...DEFAULT_FORM_FOCUS_ITEMS },
    ...form,
    focus_items: {
      ...DEFAULT_FORM_FOCUS_ITEMS,
      ...(form.focus_items || {})
    }
  };
};

window.showPage = function showPage(name, evt) {
  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  document.getElementById('page-' + name).classList.add('active');
  if (evt && evt.currentTarget) evt.currentTarget.classList.add('active');
  if (name === 'dashboard') loadDashboardPage();
  if (name === 'accounts') loadAccounts();
  if (name === 'firewall-requests') loadFirewallRequests();
  if (name === 'platform-requests') loadPlatformRequests();
  if (name === 'forms') loadFormsManagement();
  if (name === 'notifications') loadLogs();
};

window.toast = function toast(msg, type = '') {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.className = 'show ' + type;
  setTimeout(() => { el.className = ''; }, 3000);
};

window.closeModal = function closeModal(id) {
  if (window.modalIsDirty && window.modalIsDirty(id)) {
    showConfirmDialog({
      title: '⚠️ 尚未儲存',
      message: '表單內容尚未儲存，確定要關閉視窗？',
      confirmLabel: '仍要關閉',
      confirmClass: 'btn btn-danger',
      onConfirm: () => {
        const el = document.getElementById(id);
        if (el) el.classList.remove('open');
      }
    });
    return;
  }
  const el = document.getElementById(id);
  if (el) el.classList.remove('open');
};

window.showConfirmDialog = function showConfirmDialog(options) {
  const overlay = document.getElementById('confirm-overlay');
  const titleEl = document.getElementById('confirm-title');
  const msgEl = document.getElementById('confirm-msg');
  const okEl = document.getElementById('confirm-ok');
  const cancelEl = document.getElementById('confirm-cancel');
  if (!overlay || !titleEl || !msgEl || !okEl || !cancelEl) return;

  window.confirmState = {
    onConfirm: typeof options?.onConfirm === 'function' ? options.onConfirm : null,
    onCancel: typeof options?.onCancel === 'function' ? options.onCancel : null
  };

  titleEl.textContent = options?.title || '⚠️ 請再次確認';
  msgEl.textContent = options?.message || '確定要繼續執行這個操作嗎？';
  cancelEl.textContent = options?.cancelLabel || '取消';
  okEl.textContent = options?.confirmLabel || '確定';
  okEl.className = options?.confirmClass || 'btn btn-danger';
  overlay.classList.add('open');
};

window.setModalSnapshotSource = function setModalSnapshotSource(id, getter) {
  if (!id || typeof getter !== 'function') return;
  window.modalSnapshotGetters[id] = getter;
};

window.captureModalBaseline = function captureModalBaseline(id) {
  const getter = window.modalSnapshotGetters[id];
  if (!getter) return;
  window.modalSnapshotBaselines[id] = JSON.stringify(getter());
};

window.modalIsDirty = function modalIsDirty(id) {
  const getter = window.modalSnapshotGetters[id];
  if (!getter) return false;
  const current = JSON.stringify(getter());
  const baseline = window.modalSnapshotBaselines[id];
  return baseline != null && current !== baseline;
};

window.confirmOk = function confirmOk() {
  const state = window.confirmState;
  window.confirmState = null;
  document.getElementById('confirm-overlay')?.classList.remove('open');
  if (state && typeof state.onConfirm === 'function') state.onConfirm();
};

window.statusBadge = function statusBadge(s) {
  const map = { active:'badge-active', closed:'badge-closed', pending:'badge-pending', expired:'badge-expired' };
  const label = { active:'使用中', closed:'已關閉', pending:'待確認', expired:'已過期' };
  return `<span class="badge ${map[s]||'badge-default'}">${label[s]||s}</span>`;
};

window.accountTypeBadge = function accountTypeBadge(t) {
  if (t === '預設') return `<span class="badge badge-default">預設</span>`;
  if (t === '關閉帳號') return `<span class="badge badge-closed">關閉</span>`;
  return `<span class="badge" style="background:#f0fdf4;color:#166534;">${esc(t)}</span>`;
};

window.formatDate = function formatDate(d) {
  if (!d) return '—';
  return new Date(d).toLocaleDateString('zh-TW');
};

window.formatDateTime = function formatDateTime(d) {
  if (!d) return '—';
  return new Date(d).toLocaleString('zh-TW', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
};

window.countByStatus = function countByStatus(list, status) {
  return list.filter(a => a.status === status).length;
};

window.getLastUpdatedRecord = function getLastUpdatedRecord(list) {
  return list.length ? [...list].sort((a, b) => new Date(b.updated_at || b.created_at || 0) - new Date(a.updated_at || a.created_at || 0))[0] : null;
};

window.esc = function esc(s) {
  return s == null ? '' : String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
};

window.escAttr = function escAttr(s) {
  return esc(s).replace(/'/g, '&#39;');
};

window.toggleUserProfile = function toggleUserProfile() {
  profileExpanded = !profileExpanded;
  document.querySelectorAll('.user-profile').forEach(el => {
    el.classList.toggle('expanded', profileExpanded);
    const toggle = el.querySelector('.user-profile-toggle');
    if (toggle) toggle.textContent = profileExpanded ? '收合' : '展開';
  });
};

window.getCurrentUserProfile = function getCurrentUserProfile() {
  return window.currentUserProfile || {};
};

window.checkLogin = async function checkLogin() {
  const r = await fetch(API + '/api/me');
  if (!r.ok) {
    document.getElementById('login-overlay').classList.add('open');
    return;
  }
  const data = await r.json();
  window.currentUserProfile = data;
  const name = data.name || data.email.split('@')[0];
  const summaryLine = [
    data.department ? `部門：${esc(data.department)}` : '',
    data.title ? `職稱：${esc(data.title)}` : ''
  ].filter(Boolean).join(' ｜ ');
  const profileLines = [
    data.organization_name ? `組織：${esc(data.organization_name)}` : '',
    data.org_unit_path ? `OU：${esc(data.org_unit_path)}` : '',
    data.department_source ? `來源：<code>${esc(data.department_source)}</code>` : ''
  ].filter(Boolean);
  const avatar = data.picture
    ? `<img class="user-avatar" src="${escAttr(data.picture)}" alt="${escAttr(name)}">`
    : `<span class="user-avatar user-avatar-fallback">👤</span>`;
  const profileHtml = `
    ${avatar}
    <div class="user-profile-body">
      <div class="user-profile-name"><strong>${esc(name)}</strong><span class="user-profile-toggle">${profileExpanded ? '收合' : '展開'}</span></div>
      <div class="user-profile-meta">${esc(data.email || '')}</div>
      ${summaryLine ? `<div class="user-profile-meta user-profile-summary">${summaryLine}</div>` : ''}
      ${profileLines.length ? `<div class="user-profile-meta user-profile-detail">${profileLines.join('<br>')}</div>` : ''}
    </div>`;
  profileTargets.forEach(id => {
    const el = document.getElementById(id);
    if (el) {
      el.innerHTML = profileHtml;
      el.classList.toggle('expanded', profileExpanded);
      el.dataset.userName = name;
      el.dataset.userEmail = data.email || '';
      el.dataset.userDomain = data.hosted_domain || data.email_domain || '';
      el.dataset.userDepartment = data.department || '';
      el.dataset.userDepartmentSource = data.department_source || '';
      el.dataset.userTitle = data.title || '';
      el.dataset.userOrganizationName = data.organization_name || '';
      el.onclick = toggleUserProfile;
    }
  });
  await loadDashboardPage();
};

window.logout = async function logout() {
  await fetch(API + '/api/logout', { method: 'POST' });
  location.reload();
};

window.closeConfirm = function closeConfirm() {
  const state = window.confirmState;
  window.confirmState = null;
  document.getElementById('confirm-overlay')?.classList.remove('open');
  if (state && typeof state.onCancel === 'function') state.onCancel();
};

document.addEventListener('DOMContentLoaded', () => {
  checkLogin();
});
