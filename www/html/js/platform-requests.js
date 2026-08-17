(function () {
  const modalRootId = 'feature-platform-request-modal-root';
  const modalPartialPath = '/partials/platform-request-modal.html';
  const exportModalRootId = 'feature-platform-request-export-root';
  const exportModalPartialPath = '/partials/platform-request-export-modal.html';
  const exportStoragePrefix = 'isms_platform_export_reviewers';
  const fields = [
    'request_date', 'applicant_name', 'applicant_department', 'applicant_title',
    'office_phone', 'email', 'pi_name', 'system_name', 'system_alias',
    'system_purpose', 'estimated_users', 'internal_only', 'ip_restriction',
    'request_start_date', 'request_end_date', 'request_type', 'shutdown_retain_months',
    'shutdown_reason', 'environment_type', 'operating_system',
    'operating_system_other', 'disk_size', 'special_requirements',
    'domain_settings', 'other_requirements', 'backup_required',
    'backup_requirements', 'backup_reason', 'applicant_signature',
    'supervisor_signature', 'status', 'remarks'
  ];

  window.platformRequests = [];
  window.platformRequestEditingId = null;
  window.platformRequestDeleteId = null;
  window.platformRequestExportId = null;

  async function ensurePlatformRequestModalLoaded() {
    const root = document.getElementById(modalRootId);
    if (!root) return false;
    if (document.getElementById('platform-request-modal')) return true;
    const response = await fetch(modalPartialPath);
    if (!response.ok) {
      toast('申請表視窗載入失敗', 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  async function ensurePlatformRequestExportModalLoaded() {
    const root = document.getElementById(exportModalRootId);
    if (!root) return false;
    if (document.getElementById('platform-request-export-modal')) return true;
    const response = await fetch(exportModalPartialPath);
    if (!response.ok) {
      toast('匯出視窗載入失敗', 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  function getCurrentUserName() {
    return document.getElementById('user-profile-platform')?.dataset?.userName || '';
  }

  function applyCurrentUserToPlatformRequestForm() {
    const profile = window.getCurrentUserProfile ? window.getCurrentUserProfile() : {};
    const setIfEmpty = (id, value) => {
      const el = document.getElementById(id);
      if (!el || el.value.trim() || !value) return;
      el.value = value;
    };
    const fallbackName = profile.name || profile.email || '';

    setIfEmpty('pr-applicant_name', profile.name || '');
    setIfEmpty('pr-applicant_department', profile.department || '');
    setIfEmpty('pr-applicant_title', profile.title || '');
    setIfEmpty('pr-email', profile.email || '');
    setIfEmpty('pr-applicant_signature', fallbackName);
  }

  function buildExportStorageKey() {
    const user = getCurrentUserName().trim().toLowerCase();
    return `${exportStoragePrefix}:${user || 'anonymous'}`;
  }

  function readSavedExportFields() {
    try {
      const raw = window.localStorage.getItem(buildExportStorageKey());
      return raw ? JSON.parse(raw) : {};
    } catch (_) {
      return {};
    }
  }

  function saveExportFields(handlerName, managerName, reviewNotes) {
    try {
      window.localStorage.setItem(buildExportStorageKey(), JSON.stringify({
        handler_name: handlerName,
        manager_name: managerName,
        review_notes: reviewNotes
      }));
    } catch (_) {
      // Ignore storage failures and continue export flow.
    }
  }

  function todayYmd() {
    return new Date().toISOString().slice(0, 10).replace(/-/g, '');
  }

  function getPlatformRequestFormSnapshot() {
    const snapshot = {};
    fields.forEach(field => {
      snapshot[field] = document.getElementById('pr-' + field)?.value || '';
    });
    return snapshot;
  }

  window.loadPlatformRequests = async function loadPlatformRequests() {
    const r = await fetch(API + '/api/platform-requests');
    platformRequests = await r.json();
    renderPlatformRequests(platformRequests);
  };

  window.renderPlatformRequests = function renderPlatformRequests(list) {
    const tbody = document.getElementById('platform-requests-body');
    if (!tbody) return;
    if (!list || !list.length) {
      tbody.innerHTML = `<tr><td colspan="9"><div class="empty"><div class="icon">🖥️</div><p>尚無申請資料</p></div></td></tr>`;
      return;
    }
    tbody.innerHTML = list.map(item => `
      <tr>
        <td style="color:var(--text-muted);font-size:12px;">#${item.id}</td>
        <td>${esc(item.system_name)}<br><span class="hint">${esc(item.system_alias || item.environment_type)}</span></td>
        <td>${esc(item.applicant_name)}<br><span class="hint">${esc(item.applicant_department)}</span></td>
        <td>${esc(item.request_type)}<br><span class="hint">${esc(item.request_date)}</span></td>
        <td>${esc(item.operating_system)}<br><span class="hint">${esc(item.disk_size)}</span></td>
        <td>${esc(item.request_start_date)}<br><span class="hint">${esc(item.request_end_date)}</span></td>
        <td>${esc(item.email)}</td>
        <td>${statusBadge(item.status)}</td>
        <td><div class="actions"><button class="btn btn-ghost btn-sm" onclick="openPlatformRequestExport(${item.id})">⬇</button><button class="btn btn-ghost btn-sm" onclick="openPlatformRequestEdit(${item.id})">✏️</button><button class="btn btn-danger btn-sm" onclick="confirmDeletePlatformRequest(${item.id}, '${escAttr(item.system_name)}')">🗑️</button></div></td>
      </tr>`).join('');
  };

  window.openPlatformRequestExport = async function openPlatformRequestExport(id) {
    const loaded = await ensurePlatformRequestExportModalLoaded();
    if (!loaded) return;
    platformRequestExportId = id;
    const saved = readSavedExportFields();
    document.getElementById('platform-export-handler').value = saved.handler_name || '';
    document.getElementById('platform-export-manager').value = saved.manager_name || '';
    document.getElementById('platform-export-review-notes').value = saved.review_notes || '';
    document.getElementById('platform-request-export-modal').classList.add('open');
  };

  function buildPlatformRequestExportURL(format) {
    if (!platformRequestExportId) return;
    const params = new URLSearchParams();
    const handlerName = document.getElementById('platform-export-handler').value.trim();
    const managerName = document.getElementById('platform-export-manager').value.trim();
    const reviewNotes = document.getElementById('platform-export-review-notes').value.trim();

    if (handlerName) params.set('handler_name', handlerName);
    if (managerName) params.set('manager_name', managerName);
    if (reviewNotes) params.set('review_notes', reviewNotes);

    saveExportFields(handlerName, managerName, reviewNotes);
    closeModal('platform-request-export-modal');

    let url = API + '/api/platform-requests/' + platformRequestExportId + '/export-' + format;
    if ([...params].length) url += '?' + params.toString();
    return url;
  }

  window.confirmDownloadPlatformRequestDocx = function confirmDownloadPlatformRequestDocx() {
    const url = buildPlatformRequestExportURL('docx');
    if (url) window.location.href = url;
  };

  window.confirmDownloadPlatformRequestPdf = function confirmDownloadPlatformRequestPdf() {
    const url = buildPlatformRequestExportURL('pdf');
    if (url) window.location.href = url;
  };

  window.openPlatformRequestCreate = async function openPlatformRequestCreate() {
    const loaded = await ensurePlatformRequestModalLoaded();
    if (!loaded) return;
    platformRequestEditingId = null;
    document.getElementById('platform-request-modal-title').textContent = '新增系統平台申請';
    clearPlatformRequestForm();
    document.getElementById('pr-request_date').value = todayYmd();
    applyCurrentUserToPlatformRequestForm();
    setModalSnapshotSource('platform-request-modal', getPlatformRequestFormSnapshot);
    captureModalBaseline('platform-request-modal');
    document.getElementById('platform-request-modal').classList.add('open');
  };

  window.openPlatformRequestEdit = async function openPlatformRequestEdit(id) {
    const loaded = await ensurePlatformRequestModalLoaded();
    if (!loaded) return;
    platformRequestEditingId = id;
    document.getElementById('platform-request-modal-title').textContent = '編輯系統平台申請 #' + id;
    const r = await fetch(API + '/api/platform-requests/' + id);
    const data = await r.json();
    fillPlatformRequestForm(data);
    setModalSnapshotSource('platform-request-modal', getPlatformRequestFormSnapshot);
    captureModalBaseline('platform-request-modal');
    document.getElementById('platform-request-modal').classList.add('open');
  };

  window.clearPlatformRequestForm = function clearPlatformRequestForm() {
    fields.forEach(field => {
      const el = document.getElementById('pr-' + field);
      if (el) el.value = '';
    });
    document.getElementById('pr-status').value = 'active';
    document.getElementById('pr-internal_only').value = '是';
    document.getElementById('pr-request_type').value = '上架新增';
    document.getElementById('pr-environment_type').value = '正式環境';
    document.getElementById('pr-operating_system').value = 'Rocky 9';
    document.getElementById('pr-backup_required').value = '是';
  };

  window.fillPlatformRequestForm = function fillPlatformRequestForm(item) {
    fields.forEach(field => {
      const el = document.getElementById('pr-' + field);
      if (el) el.value = item[field] || '';
    });
  };

  window.savePlatformRequest = async function savePlatformRequest() {
    const payload = {};
    fields.forEach(field => {
      const el = document.getElementById('pr-' + field);
      payload[field] = el ? el.value.trim() : '';
    });
    if (!payload.system_name || !payload.applicant_name) {
      return toast('請填寫系統名稱與申請人姓名', 'error');
    }
    const url = platformRequestEditingId ? API + '/api/platform-requests/' + platformRequestEditingId : API + '/api/platform-requests';
    const method = platformRequestEditingId ? 'PUT' : 'POST';
    const r = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await r.json();
    if (!r.ok) return toast('儲存失敗：' + (data.error || '未知錯誤'), 'error');
    captureModalBaseline('platform-request-modal');
    closeModal('platform-request-modal');
    toast(platformRequestEditingId ? '申請資料已更新' : '申請資料已新增', 'success');
    await Promise.all([loadPlatformRequests(), loadDashboardPage()]);
  };

  window.confirmDeletePlatformRequest = function confirmDeletePlatformRequest(id, name) {
    platformRequestDeleteId = id;
    showConfirmDialog({
      title: '⚠️ 確認刪除',
      message: `確定要刪除申請「${name}」(#${id})？此操作無法復原。`,
      confirmLabel: '確定刪除',
      confirmClass: 'btn btn-danger',
      onConfirm: doDeletePlatformRequest
    });
  };

  window.doDeletePlatformRequest = async function doDeletePlatformRequest() {
    const r = await fetch(API + '/api/platform-requests/' + platformRequestDeleteId, { method: 'DELETE' });
    closeConfirm();
    if (!r.ok) return toast('刪除申請失敗', 'error');
    toast('申請資料已刪除', 'success');
    await Promise.all([loadPlatformRequests(), loadDashboardPage()]);
  };
})();
