(function () {
  const modalRootId = 'feature-asset-inventory-modal-root';
  const modalPartialPath = '/partials/asset-inventory-modal.html';
  const fields = [
    'system_name', 'asset_code', 'asset_type', 'asset_name', 'vendor_name',
    'is_core_asset', 'has_national_security_concern', 'asset_description', 'quantity',
    'os_config_baseline', 'browser_config_baseline', 'network_config_baseline',
    'application_config_baseline', 'other_config_baseline', 'config_exception_code',
    'manager_department', 'user_department', 'location', 'confidentiality',
    'integrity', 'availability', 'asset_value', 'legal_compliance',
    'protection_level', 'mtpd', 'rto', 'rpo', 'status', 'remarks'
  ];

  window.assetInventoryRecords = [];
  window.assetInventoryEditingId = null;
  window.assetInventoryDeleteId = null;

  window.buildAssetInventoryQueryParams = function buildAssetInventoryQueryParams(status, q, assetType) {
    const params = new URLSearchParams();
    if (q) params.set('q', q);
    if (status && status !== 'all') params.set('status', status);
    if (assetType && assetType !== 'all') params.set('asset_type', assetType);
    return params;
  };

  async function ensureAssetInventoryModalLoaded() {
    const root = document.getElementById(modalRootId);
    if (!root) return false;
    if (document.getElementById('asset-inventory-modal')) return true;
    const response = await fetch(modalPartialPath);
    if (!response.ok) {
      toast('資訊資產視窗載入失敗', 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  function getAssetInventoryFormSnapshot() {
    const snapshot = {};
    fields.forEach(field => {
      snapshot[field] = document.getElementById('ai-' + field)?.value || '';
    });
    return snapshot;
  }

  window.loadAssetInventory = async function loadAssetInventory() {
    const status = document.getElementById('asset-status-filter')?.value || 'all';
    const q = document.getElementById('asset-search-input')?.value.trim() || '';
    const assetType = document.getElementById('asset-type-filter')?.value || 'all';
    let url = API + '/api/asset-inventory';
    const params = buildAssetInventoryQueryParams(status, q, assetType);
    if ([...params].length) url += '?' + params.toString();
    const r = await fetch(url);
    assetInventoryRecords = await r.json();
    renderAssetInventory(assetInventoryRecords);
  };

  window.debounceAssetInventorySearch = function debounceAssetInventorySearch() {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(loadAssetInventory, 350);
  };

  window.renderAssetInventory = function renderAssetInventory(list) {
    const tbody = document.getElementById('asset-inventory-body');
    if (!tbody) return;
    if (!list || !list.length) {
      tbody.innerHTML = `<tr><td colspan="10"><div class="empty"><div class="icon">📦</div><p>尚無資訊資產資料</p></div></td></tr>`;
      return;
    }
    tbody.innerHTML = list.map(item => `
      <tr>
        <td style="color:var(--text-muted);font-size:12px;">#${item.id}</td>
        <td>${esc(item.system_name)}<br><span class="hint">${esc(item.asset_code || '—')}</span></td>
        <td>${esc(item.asset_type)}<br><span class="hint">${esc(item.asset_name)}</span></td>
        <td>${esc(item.manager_department)}<br><span class="hint">${esc(item.user_department)}</span></td>
        <td>${esc(item.location)}</td>
        <td>${esc(item.confidentiality || '—')} / ${esc(item.integrity || '—')} / ${esc(item.availability || '—')}</td>
        <td>${esc(item.asset_value || '—')} / ${esc(item.protection_level || '—')}</td>
        <td>${esc(item.quantity)}</td>
        <td>${statusBadge(item.status)}</td>
        <td><div class="actions"><button class="btn btn-ghost btn-sm" onclick="openAssetInventoryEdit(${item.id})">✏️</button><button class="btn btn-danger btn-sm" onclick="confirmDeleteAssetInventory(${item.id}, '${escAttr(item.asset_name || item.system_name)}')">🗑️</button></div></td>
      </tr>`).join('');
  };

  window.openAssetInventoryCreate = async function openAssetInventoryCreate() {
    const loaded = await ensureAssetInventoryModalLoaded();
    if (!loaded) return;
    assetInventoryEditingId = null;
    document.getElementById('asset-inventory-modal-title').textContent = '新增資訊資產';
    clearAssetInventoryForm();
    setModalSnapshotSource('asset-inventory-modal', getAssetInventoryFormSnapshot);
    captureModalBaseline('asset-inventory-modal');
    document.getElementById('asset-inventory-modal').classList.add('open');
  };

  window.openAssetInventoryEdit = async function openAssetInventoryEdit(id) {
    const loaded = await ensureAssetInventoryModalLoaded();
    if (!loaded) return;
    assetInventoryEditingId = id;
    document.getElementById('asset-inventory-modal-title').textContent = '編輯資訊資產 #' + id;
    const r = await fetch(API + '/api/asset-inventory/' + id);
    const data = await r.json();
    fillAssetInventoryForm(data);
    setModalSnapshotSource('asset-inventory-modal', getAssetInventoryFormSnapshot);
    captureModalBaseline('asset-inventory-modal');
    document.getElementById('asset-inventory-modal').classList.add('open');
  };

  window.clearAssetInventoryForm = function clearAssetInventoryForm() {
    fields.forEach(field => {
      const el = document.getElementById('ai-' + field);
      if (el) el.value = '';
    });
    document.getElementById('ai-asset_type').value = '實體類';
    document.getElementById('ai-is_core_asset').value = '否';
    document.getElementById('ai-has_national_security_concern').value = '否';
    document.getElementById('ai-quantity').value = '1';
    document.getElementById('ai-status').value = 'active';
  };

  window.fillAssetInventoryForm = function fillAssetInventoryForm(item) {
    fields.forEach(field => {
      const el = document.getElementById('ai-' + field);
      if (el) el.value = item[field] || '';
    });
  };

  window.saveAssetInventory = async function saveAssetInventory() {
    const payload = {};
    fields.forEach(field => {
      const el = document.getElementById('ai-' + field);
      payload[field] = el ? el.value.trim() : '';
    });
    if (!payload.system_name || !payload.asset_type || !payload.asset_name) {
      return toast('請填寫資通系統名稱、資產類別與資產名稱', 'error');
    }
    const url = assetInventoryEditingId ? API + '/api/asset-inventory/' + assetInventoryEditingId : API + '/api/asset-inventory';
    const method = assetInventoryEditingId ? 'PUT' : 'POST';
    const r = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await r.json();
    if (!r.ok) return toast('儲存失敗：' + (data.error || '未知錯誤'), 'error');
    captureModalBaseline('asset-inventory-modal');
    closeModal('asset-inventory-modal');
    toast(assetInventoryEditingId ? '資訊資產已更新' : '資訊資產已新增', 'success');
    await Promise.all([loadAssetInventory(), loadDashboardPage()]);
  };

  window.confirmDeleteAssetInventory = function confirmDeleteAssetInventory(id, name) {
    assetInventoryDeleteId = id;
    showConfirmDialog({
      title: '⚠️ 確認刪除',
      message: `確定要刪除資訊資產「${name}」(#${id})？此操作無法復原。`,
      confirmLabel: '確定刪除',
      confirmClass: 'btn btn-danger',
      onConfirm: doDeleteAssetInventory
    });
  };

  window.doDeleteAssetInventory = async function doDeleteAssetInventory() {
    const r = await fetch(API + '/api/asset-inventory/' + assetInventoryDeleteId, { method: 'DELETE' });
    closeConfirm();
    if (!r.ok) return toast('刪除資訊資產失敗', 'error');
    toast('資訊資產已刪除', 'success');
    await Promise.all([loadAssetInventory(), loadDashboardPage()]);
  };

  window.downloadAssetInventoryXlsx = function downloadAssetInventoryXlsx() {
    downloadAssetInventoryFile('xlsx');
  };

  window.downloadAssetInventoryDocx = function downloadAssetInventoryDocx() {
    downloadAssetInventoryFile('docx');
  };

  window.downloadAssetInventoryPdf = function downloadAssetInventoryPdf() {
    downloadAssetInventoryFile('pdf');
  };

  function downloadAssetInventoryFile(format) {
    const status = document.getElementById('asset-status-filter')?.value || 'all';
    const q = document.getElementById('asset-search-input')?.value.trim() || '';
    const assetType = document.getElementById('asset-type-filter')?.value || 'all';
    let url = API + '/api/asset-inventory/export-' + format;
    const params = buildAssetInventoryQueryParams(status, q, assetType);
    if ([...params].length) url += '?' + params.toString();
    window.location.href = url;
  }

  window.triggerAssetInventoryImport = function triggerAssetInventoryImport() {
    document.getElementById('asset-import-file')?.click();
  };

  window.importAssetInventoryXlsx = async function importAssetInventoryXlsx(input) {
    const file = input?.files?.[0];
    if (!file) return;
    const formData = new FormData();
    formData.append('file', file);
    const r = await fetch(API + '/api/asset-inventory/import-xlsx', {
      method: 'POST',
      body: formData
    });
    const data = await r.json();
    input.value = '';
    if (!r.ok) return toast('匯入失敗：' + (data.error || '未知錯誤'), 'error');
    toast(`匯入完成：新增 ${data.imported || 0} 筆，略過 ${data.skipped || 0} 筆`, 'success');
    await Promise.all([loadAssetInventory(), loadDashboardPage()]);
  };
})();
