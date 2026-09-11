(function () {
  const modalRootId = 'feature-application-change-modal-root';
  const modalPartialPath = '/partials/application-change-request-modal.html';
  const fields = [
    'suggestor', 'form_date', 'approver', 'system_name', 'feature_name',
    'is_required_feature', 'is_major_impact', 'expected_online_date',
    'background_description', 'existing_security_measures', 'security_scope_involved',
    'access_control_measures', 'audit_measures', 'continuity_measures',
    'identification_measures', 'system_acquisition_measures', 'communication_measures',
    'integrity_measures', 'capacity_management', 'capacity_change_description',
    'other_security_measures', 'other_security_description', 'information_service_opinion',
    'meeting_time', 'meeting_decision', 'rejection_reason', 'coordinating_staff',
    'coordination_date', 'coordination_approver', 'status', 'remarks'
  ];
  const dateFields = ['form_date', 'expected_online_date', 'coordination_date'];

  window.applicationChangeRequests = [];
  window.applicationChangeEditingId = null;
  window.applicationChangeDeleteId = null;

  async function ensureApplicationChangeModalLoaded() {
    const root = document.getElementById(modalRootId);
    if (!root) return false;
    if (document.getElementById('application-change-request-modal')) return true;
    const response = await fetch(modalPartialPath);
    if (!response.ok) {
      toast('建議單視窗載入失敗', 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  function applyCurrentUserToApplicationChangeForm() {
    const profile = window.getCurrentUserProfile ? window.getCurrentUserProfile() : {};
    const el = document.getElementById('acr-suggestor');
    if (el && !el.value.trim()) el.value = profile.name || profile.email || '';
  }

  function getApplicationChangeFormSnapshot() {
    const snapshot = {};
    fields.forEach(field => {
      snapshot[field] = document.getElementById('acr-' + field)?.value || '';
    });
    return snapshot;
  }

  window.loadApplicationChangeRequests = async function loadApplicationChangeRequests() {
    const r = await fetch(API + '/api/application-change-requests');
    applicationChangeRequests = await r.json();
    renderApplicationChangeRequests(applicationChangeRequests);
  };

  window.renderApplicationChangeRequests = function renderApplicationChangeRequests(list) {
    const tbody = document.getElementById('application-change-requests-body');
    if (!tbody) return;
    if (!list || !list.length) {
      tbody.innerHTML = `<tr><td colspan="8"><div class="empty"><p>尚無功能需求更新建議資料</p></div></td></tr>`;
      return;
    }
    tbody.innerHTML = list.map(item => `
      <tr>
        <td style="color:var(--text-muted);font-size:12px;">#${item.id}</td>
        <td>${esc(item.system_name)}<br><span class="hint">${esc(item.feature_name)}</span></td>
        <td>${esc(item.suggestor)}<br><span class="hint">${esc(formatDate(item.form_date))}</span></td>
        <td>制式：${esc(item.is_required_feature || '未選')}<br><span class="hint">重大：${esc(item.is_major_impact || '未選')}</span></td>
        <td>${esc(formatDate(item.expected_online_date))}</td>
        <td>${esc(item.information_service_opinion || '未選擇')}</td>
        <td>${statusBadge(item.status)}</td>
        <td><div class="actions"><button class="btn btn-ghost btn-sm" onclick="openApplicationChangeEdit(${item.id})">✏️</button><button class="btn btn-danger btn-sm" onclick="confirmDeleteApplicationChange(${item.id}, '${escAttr(item.feature_name || item.system_name)}')">🗑️</button></div></td>
      </tr>`).join('');
  };

  window.openApplicationChangeCreate = async function openApplicationChangeCreate() {
    const loaded = await ensureApplicationChangeModalLoaded();
    if (!loaded) return;
    applicationChangeEditingId = null;
    document.getElementById('application-change-request-modal-title').textContent = '新增功能需求更新建議';
    clearApplicationChangeForm();
    initDateInputs(document.getElementById('application-change-request-modal'));
    document.getElementById('acr-form_date').value = todayDateInputValue();
    applyCurrentUserToApplicationChangeForm();
    setModalSnapshotSource('application-change-request-modal', getApplicationChangeFormSnapshot);
    captureModalBaseline('application-change-request-modal');
    document.getElementById('application-change-request-modal').classList.add('open');
  };

  window.openApplicationChangeEdit = async function openApplicationChangeEdit(id) {
    const loaded = await ensureApplicationChangeModalLoaded();
    if (!loaded) return;
    applicationChangeEditingId = id;
    document.getElementById('application-change-request-modal-title').textContent = '編輯功能需求更新建議 #' + id;
    const r = await fetch(API + '/api/application-change-requests/' + id);
    const data = await r.json();
    fillApplicationChangeForm(data);
    initDateInputs(document.getElementById('application-change-request-modal'));
    setModalSnapshotSource('application-change-request-modal', getApplicationChangeFormSnapshot);
    captureModalBaseline('application-change-request-modal');
    document.getElementById('application-change-request-modal').classList.add('open');
  };

  window.clearApplicationChangeForm = function clearApplicationChangeForm() {
    fields.forEach(field => {
      const el = document.getElementById('acr-' + field);
      if (el) el.value = '';
    });
    document.getElementById('acr-status').value = 'active';
  };

  window.fillApplicationChangeForm = function fillApplicationChangeForm(item) {
    fields.forEach(field => {
      const el = document.getElementById('acr-' + field);
      if (el) el.value = dateFields.includes(field) ? normalizeDateValue(item[field] || '') : (item[field] || '');
    });
  };

  window.saveApplicationChangeRequest = async function saveApplicationChangeRequest() {
    const payload = {};
    fields.forEach(field => {
      const el = document.getElementById('acr-' + field);
      payload[field] = el ? el.value.trim() : '';
    });
    if (!payload.system_name || !payload.feature_name || !payload.suggestor) {
      return toast('請填寫應用系統、功能需求名稱與建議者', 'error');
    }
    const url = applicationChangeEditingId ? API + '/api/application-change-requests/' + applicationChangeEditingId : API + '/api/application-change-requests';
    const method = applicationChangeEditingId ? 'PUT' : 'POST';
    const r = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await r.json();
    if (!r.ok) return toast('儲存失敗：' + (data.error || '未知錯誤'), 'error');
    captureModalBaseline('application-change-request-modal');
    closeModal('application-change-request-modal');
    toast(applicationChangeEditingId ? '建議單已更新' : '建議單已新增', 'success');
    await Promise.all([loadApplicationChangeRequests(), loadDashboardPage()]);
  };

  window.confirmDeleteApplicationChange = function confirmDeleteApplicationChange(id, name) {
    applicationChangeDeleteId = id;
    showConfirmDialog({
      title: '⚠️ 確認刪除',
      message: `確定要刪除建議「${name}」(#${id})？此操作無法復原。`,
      confirmLabel: '確定刪除',
      confirmClass: 'btn btn-danger',
      onConfirm: doDeleteApplicationChange
    });
  };

  window.doDeleteApplicationChange = async function doDeleteApplicationChange() {
    const r = await fetch(API + '/api/application-change-requests/' + applicationChangeDeleteId, { method: 'DELETE' });
    closeConfirm();
    if (!r.ok) return toast('刪除建議單失敗', 'error');
    toast('建議單已刪除', 'success');
    await Promise.all([loadApplicationChangeRequests(), loadDashboardPage()]);
  };
})();
