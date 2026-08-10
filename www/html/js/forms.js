(function () {
  async function ensureFormConfigModalLoaded() {
    const root = document.getElementById('feature-form-config-modal-root');
    if (!root) return false;
    if (document.getElementById('form-config-modal')) return true;

    const response = await fetch('/partials/form-config-modal.html');
    if (!response.ok) {
      toast('表單設定視窗載入失敗', 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  window.loadFormsManagement = async function loadFormsManagement() {
    await loadDashboardConfig();
    renderFormsManagement();
  };

  window.renderFormsManagement = function renderFormsManagement() {
    const rows = dashboardForms.map((form, index) => `
      <tr>
        <td>${index + 1}</td>
        <td><span class="mono">${esc(form.key)}</span></td>
        <td>${esc(form.code)}<br><span class="hint">${esc(form.short_code)}</span></td>
        <td>${esc(form.name)}<br><span class="hint">${esc(form.description || '')}</span></td>
        <td>${providerLabel(form.provider_key)}</td>
        <td>${form.enabled ? '<span class="badge badge-active">啟用</span>' : '<span class="badge badge-expired">停用</span>'}</td>
        <td>
          <div class="order-actions">
            <button class="btn btn-ghost btn-sm" onclick="moveDashboardForm(${form.id}, -1)" ${index === 0 ? 'disabled' : ''}>↑</button>
            <button class="btn btn-ghost btn-sm" onclick="moveDashboardForm(${form.id}, 1)" ${index === dashboardForms.length - 1 ? 'disabled' : ''}>↓</button>
            <button class="btn btn-ghost btn-sm" onclick="openFormEdit(${form.id})">✏️</button>
            <button class="btn btn-danger btn-sm" onclick="deleteDashboardForm(${form.id}, '${escAttr(form.name)}')">🗑️</button>
          </div>
        </td>
      </tr>`).join('');
    document.getElementById('forms-management-list').innerHTML = `
      <div class="table-wrap">
        <table>
          <thead><tr><th>順序</th><th>Registry Key</th><th>表單代碼</th><th>名稱 / 說明</th><th>Data Provider</th><th>狀態</th><th width="220">操作</th></tr></thead>
          <tbody>${rows || `<tr><td colspan="7"><div class="empty"><div class="icon">🗂️</div><p>尚無表單設定</p></div></td></tr>`}</tbody>
        </table>
      </div>`;
  };

  window.providerLabel = function providerLabel(key) {
    const provider = dashboardProviders.find(p => p.key === key);
    return provider ? `${provider.label}<br><span class="hint mono">${provider.key}</span>` : `<span class="mono">${esc(key)}</span>`;
  };

  window.renderProviderOptions = function renderProviderOptions() {
    document.getElementById('ff-provider_key').innerHTML = dashboardProviders.map(p => `<option value="${escAttr(p.key)}">${esc(p.label)} (${esc(p.key)})</option>`).join('');
  };

  window.openFormCreate = async function openFormCreate() {
    const loaded = await ensureFormConfigModalLoaded();
    if (!loaded) return;
    formEditingId = null;
    renderProviderOptions();
    document.getElementById('form-config-title').textContent = '新增表單設定';
    clearDashboardFormForm();
    document.getElementById('ff-display_order').value = dashboardForms.length + 1;
    document.getElementById('form-config-modal').classList.add('open');
  };

  window.openFormEdit = async function openFormEdit(id) {
    const loaded = await ensureFormConfigModalLoaded();
    if (!loaded) return;
    const form = dashboardForms.find(x => x.id === id);
    if (!form) return;
    formEditingId = id;
    renderProviderOptions();
    document.getElementById('form-config-title').textContent = '編輯表單設定';
    fillDashboardFormForm(form);
    document.getElementById('form-config-modal').classList.add('open');
  };

  window.clearDashboardFormForm = function clearDashboardFormForm() {
    ['key','code','short_code','name','description','detail_title','empty_text','status_normal_text','status_needs_attention_text','active_title','active_meta','pending_title','pending_meta','closed_title','closed_meta','recent_title'].forEach(id => {
      const el = document.getElementById('ff-' + id);
      if (el) el.value = '';
    });
    document.getElementById('ff-enabled').value = 'true';
    if (dashboardProviders[0]) document.getElementById('ff-provider_key').value = dashboardProviders[0].key;
  };

  window.fillDashboardFormForm = function fillDashboardFormForm(form) {
    document.getElementById('ff-key').value = form.key || '';
    document.getElementById('ff-code').value = form.code || '';
    document.getElementById('ff-short_code').value = form.short_code || '';
    document.getElementById('ff-name').value = form.name || '';
    document.getElementById('ff-description').value = form.description || '';
    document.getElementById('ff-detail_title').value = form.detail_title || '';
    document.getElementById('ff-empty_text').value = form.empty_text || '';
    document.getElementById('ff-status_normal_text').value = form.status_normal_text || '';
    document.getElementById('ff-status_needs_attention_text').value = form.status_needs_attention_text || '';
    document.getElementById('ff-provider_key').value = form.provider_key || '';
    document.getElementById('ff-display_order').value = form.display_order || 1;
    document.getElementById('ff-enabled').value = String(!!form.enabled);
    const focus = form.focus_items || DEFAULT_FORM_FOCUS_ITEMS;
    document.getElementById('ff-active_title').value = focus.active_title || '';
    document.getElementById('ff-active_meta').value = focus.active_meta || '';
    document.getElementById('ff-pending_title').value = focus.pending_title || '';
    document.getElementById('ff-pending_meta').value = focus.pending_meta || '';
    document.getElementById('ff-closed_title').value = focus.closed_title || '';
    document.getElementById('ff-closed_meta').value = focus.closed_meta || '';
    document.getElementById('ff-recent_title').value = focus.recent_title || '';
  };

  window.saveDashboardForm = async function saveDashboardForm() {
    const payload = {
      key: document.getElementById('ff-key').value.trim(),
      code: document.getElementById('ff-code').value.trim(),
      short_code: document.getElementById('ff-short_code').value.trim(),
      name: document.getElementById('ff-name').value.trim(),
      description: document.getElementById('ff-description').value.trim(),
      detail_title: document.getElementById('ff-detail_title').value.trim(),
      empty_text: document.getElementById('ff-empty_text').value.trim(),
      status_normal_text: document.getElementById('ff-status_normal_text').value.trim(),
      status_needs_attention_text: document.getElementById('ff-status_needs_attention_text').value.trim(),
      provider_key: document.getElementById('ff-provider_key').value,
      display_order: Number(document.getElementById('ff-display_order').value || 1),
      enabled: document.getElementById('ff-enabled').value === 'true',
      focus_items: {
        active_title: document.getElementById('ff-active_title').value.trim(),
        active_meta: document.getElementById('ff-active_meta').value.trim(),
        pending_title: document.getElementById('ff-pending_title').value.trim(),
        pending_meta: document.getElementById('ff-pending_meta').value.trim(),
        closed_title: document.getElementById('ff-closed_title').value.trim(),
        closed_meta: document.getElementById('ff-closed_meta').value.trim(),
        recent_title: document.getElementById('ff-recent_title').value.trim()
      }
    };
    if (!payload.key || !payload.code || !payload.short_code || !payload.name) return toast('請填寫表單必要欄位', 'error');
    const url = formEditingId ? API + '/api/dashboard-forms/' + formEditingId : API + '/api/dashboard-forms';
    const method = formEditingId ? 'PUT' : 'POST';
    const r = await fetch(url, { method, headers: {'Content-Type':'application/json'}, body: JSON.stringify(payload) });
    const data = await r.json();
    if (!r.ok) return toast('表單設定儲存失敗：' + (data.error || '未知錯誤'), 'error');
    closeModal('form-config-modal');
    toast(formEditingId ? '表單設定已更新' : '表單設定已新增', 'success');
    await Promise.all([loadFormsManagement(), loadDashboardPage()]);
  };

  window.deleteDashboardForm = async function deleteDashboardForm(id, name) {
    if (!confirm(`確定要刪除表單「${name}」？`)) return;
    const r = await fetch(API + '/api/dashboard-forms/' + id, { method: 'DELETE' });
    if (!r.ok) return toast('刪除表單失敗', 'error');
    toast('表單設定已刪除', 'success');
    await Promise.all([loadFormsManagement(), loadDashboardPage()]);
  };

  window.moveDashboardForm = async function moveDashboardForm(id, delta) {
    const index = dashboardForms.findIndex(f => f.id === id);
    const target = index + delta;
    if (index < 0 || target < 0 || target >= dashboardForms.length) return;
    const reordered = [...dashboardForms];
    const [item] = reordered.splice(index, 1);
    reordered.splice(target, 0, item);
    const ids = reordered.map(f => f.id);
    const r = await fetch(API + '/api/dashboard-forms/reorder', {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify({ ids })
    });
    if (!r.ok) return toast('更新順序失敗', 'error');
    await Promise.all([loadFormsManagement(), loadDashboardPage()]);
  };
})();
