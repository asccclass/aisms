(function () {
  const modalRootId = 'feature-protection-baseline-modal-root';
  const modalPartialPath = '/partials/protection-baseline-modal.html';
  const headerFields = ['system_name', 'security_level', 'filled_by', 'form_date', 'reviewer', 'review_date', 'status', 'remarks'];

  window.protectionBaselineRecords = [];
  window.protectionBaselineEditingId = null;
  window.protectionBaselineDeleteId = null;

  async function ensureProtectionBaselineModalLoaded() {
    const root = document.getElementById(modalRootId);
    if (!root) return false;
    if (document.getElementById('protection-baseline-modal')) return true;
    const response = await fetch(modalPartialPath);
    if (!response.ok) {
      toast('04-069 視窗載入失敗', 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  function getProtectionBaselineSnapshot() {
    return JSON.stringify(buildProtectionBaselinePayload());
  }

  function buildProtectionBaselinePayload() {
    const payload = {};
    headerFields.forEach(field => {
      payload[field] = document.getElementById('pb-' + field)?.value.trim() || '';
    });
    payload.controls = [...document.querySelectorAll('#pb-controls-body tr')].map(tr => ({
      item_no: Number(tr.dataset.itemNo || 0),
      domain_name: tr.dataset.domainName || '',
      control_category: tr.dataset.controlCategory || '',
      requirement_level: tr.dataset.requirementLevel || '',
      control_description: tr.dataset.controlDescription || '',
      measure_notes: tr.dataset.measureNotes || '',
      applies: tr.querySelector('.pb-applies')?.value || '',
      implementation_notes: tr.querySelector('.pb-implementation')?.value.trim() || '',
      compliance: tr.querySelector('.pb-compliance')?.value || '',
      finding: tr.querySelector('.pb-finding')?.value.trim() || '',
      remarks: tr.querySelector('.pb-remarks')?.value.trim() || (tr.dataset.remarks || '')
    }));
    return payload;
  }

  function renderControls(controls) {
    const tbody = document.getElementById('pb-controls-body');
    tbody.innerHTML = controls.map(item => `
      <tr
        data-item-no="${item.item_no}"
        data-domain-name="${escAttr(item.domain_name)}"
        data-control-category="${escAttr(item.control_category)}"
        data-requirement-level="${escAttr(item.requirement_level)}"
        data-control-description="${escAttr(item.control_description)}"
        data-measure-notes="${escAttr(item.measure_notes || '')}"
        data-remarks="${escAttr(item.remarks || '')}">
        <td>${item.item_no}</td>
        <td>${esc(item.domain_name)}<br><span class="hint">${esc(item.control_category)}</span></td>
        <td>${esc(item.requirement_level)}</td>
        <td>${esc(item.control_description)}${item.measure_notes ? `<div class="hint" style="margin-top:4px;">${esc(item.measure_notes)}</div>` : ''}</td>
        <td><select class="form-control pb-applies"><option value="">未填</option><option value="Y" ${item.applies === 'Y' ? 'selected' : ''}>Y</option><option value="N" ${item.applies === 'N' ? 'selected' : ''}>N</option><option value="不適用" ${item.applies === '不適用' ? 'selected' : ''}>不適用</option></select></td>
        <td><textarea class="form-control pb-implementation" rows="2">${esc(item.implementation_notes || '')}</textarea></td>
        <td><select class="form-control pb-compliance"><option value="">未填</option><option value="符合" ${item.compliance === '符合' ? 'selected' : ''}>符合</option><option value="不符合" ${item.compliance === '不符合' ? 'selected' : ''}>不符合</option><option value="不適用" ${item.compliance === '不適用' ? 'selected' : ''}>不適用</option></select></td>
        <td><textarea class="form-control pb-finding" rows="2">${esc(item.finding || '')}</textarea></td>
        <td><textarea class="form-control pb-remarks" rows="2">${esc(item.remarks || '')}</textarea></td>
      </tr>`).join('');
  }

  window.loadProtectionBaselines = async function loadProtectionBaselines() {
    const r = await fetch(API + '/api/protection-baselines');
    protectionBaselineRecords = await r.json();
    renderProtectionBaselines(protectionBaselineRecords);
  };

  window.renderProtectionBaselines = function renderProtectionBaselines(list) {
    const tbody = document.getElementById('protection-baselines-body');
    if (!tbody) return;
    if (!list || !list.length) {
      tbody.innerHTML = `<tr><td colspan="9"><div class="empty"><div class="icon">🛡️</div><p>尚無 04-069 資料</p></div></td></tr>`;
      return;
    }
    tbody.innerHTML = list.map(item => `
      <tr>
        <td style="color:var(--text-muted);font-size:12px;">#${item.id}</td>
        <td>${esc(item.system_name)}<br><span class="hint">安全等級：${esc(item.security_level || '—')}</span></td>
        <td>${esc(item.filled_by || '—')}<br><span class="hint">${esc(formatDate(item.form_date))}</span></td>
        <td>${esc(item.reviewer || '—')}<br><span class="hint">${esc(formatDate(item.review_date))}</span></td>
        <td>${item.applied_count} / ${item.total_controls}</td>
        <td>${item.compliant_count} / ${item.total_controls}</td>
        <td>${statusBadge(item.status)}</td>
        <td>${esc(item.creator || '—')}</td>
        <td><div class="actions"><button class="btn btn-ghost btn-sm" onclick="downloadProtectionBaselineXlsx(${item.id})">⬇</button><button class="btn btn-ghost btn-sm" onclick="openProtectionBaselineEdit(${item.id})">✏️</button><button class="btn btn-danger btn-sm" onclick="confirmDeleteProtectionBaseline(${item.id}, '${escAttr(item.system_name)}')">🗑️</button></div></td>
      </tr>`).join('');
  };

  window.downloadProtectionBaselineXlsx = function downloadProtectionBaselineXlsx(id) {
    window.location.href = API + '/api/protection-baselines/' + id + '/export-xlsx';
  };

  window.openProtectionBaselineCreate = async function openProtectionBaselineCreate() {
    const loaded = await ensureProtectionBaselineModalLoaded();
    if (!loaded) return;
    protectionBaselineEditingId = null;
    document.getElementById('protection-baseline-modal-title').textContent = '新增 04-069 資料';
    clearProtectionBaselineForm();
    const templateRes = await fetch(API + '/api/protection-baselines/template');
    const templateControls = templateRes.ok ? await templateRes.json() : [];
    renderControls(templateControls);
    initDateInputs(document.getElementById('protection-baseline-modal'));
    document.getElementById('pb-form_date').value = todayDateInputValue();
    setModalSnapshotSource('protection-baseline-modal', getProtectionBaselineSnapshot);
    captureModalBaseline('protection-baseline-modal');
    document.getElementById('protection-baseline-modal').classList.add('open');
  };

  window.openProtectionBaselineEdit = async function openProtectionBaselineEdit(id) {
    const loaded = await ensureProtectionBaselineModalLoaded();
    if (!loaded) return;
    protectionBaselineEditingId = id;
    document.getElementById('protection-baseline-modal-title').textContent = '編輯 04-069 #' + id;
    const r = await fetch(API + '/api/protection-baselines/' + id);
    const data = await r.json();
    fillProtectionBaselineForm(data);
    initDateInputs(document.getElementById('protection-baseline-modal'));
    setModalSnapshotSource('protection-baseline-modal', getProtectionBaselineSnapshot);
    captureModalBaseline('protection-baseline-modal');
    document.getElementById('protection-baseline-modal').classList.add('open');
  };

  window.clearProtectionBaselineForm = function clearProtectionBaselineForm() {
    headerFields.forEach(field => {
      const el = document.getElementById('pb-' + field);
      if (el) el.value = '';
    });
    document.getElementById('pb-security_level').value = '普';
    document.getElementById('pb-status').value = 'active';
    renderControls([]);
  };

  window.fillProtectionBaselineForm = function fillProtectionBaselineForm(item) {
    headerFields.forEach(field => {
      const el = document.getElementById('pb-' + field);
      if (el) el.value = ['form_date', 'review_date'].includes(field) ? normalizeDateValue(item[field] || '') : (item[field] || '');
    });
    renderControls(item.controls || []);
  };

  window.saveProtectionBaseline = async function saveProtectionBaseline() {
    const payload = buildProtectionBaselinePayload();
    if (!payload.system_name || !payload.security_level || !payload.filled_by || !payload.form_date) {
      return toast('請填寫系統名稱、安全等級、填寫人與填表日期', 'error');
    }
    const url = protectionBaselineEditingId ? API + '/api/protection-baselines/' + protectionBaselineEditingId : API + '/api/protection-baselines';
    const method = protectionBaselineEditingId ? 'PUT' : 'POST';
    const r = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await r.json();
    if (!r.ok) return toast('儲存失敗：' + (data.error || '未知錯誤'), 'error');
    captureModalBaseline('protection-baseline-modal');
    closeModal('protection-baseline-modal');
    toast(protectionBaselineEditingId ? '04-069 資料已更新' : '04-069 資料已新增', 'success');
    await Promise.all([loadProtectionBaselines(), loadDashboardPage()]);
  };

  window.confirmDeleteProtectionBaseline = function confirmDeleteProtectionBaseline(id, name) {
    protectionBaselineDeleteId = id;
    showConfirmDialog({
      title: '⚠️ 確認刪除',
      message: `確定要刪除 04-069「${name}」(#${id})？此操作無法復原。`,
      confirmLabel: '確定刪除',
      confirmClass: 'btn btn-danger',
      onConfirm: doDeleteProtectionBaseline
    });
  };

  window.doDeleteProtectionBaseline = async function doDeleteProtectionBaseline() {
    const r = await fetch(API + '/api/protection-baselines/' + protectionBaselineDeleteId, { method: 'DELETE' });
    closeConfirm();
    if (!r.ok) return toast('刪除 04-069 失敗', 'error');
    toast('04-069 資料已刪除', 'success');
    await Promise.all([loadProtectionBaselines(), loadDashboardPage()]);
  };
})();
