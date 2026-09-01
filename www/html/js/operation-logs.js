(function () {
  window.runCreatorBackfill = async function runCreatorBackfill() {
    showConfirmDialog({
      title: '⚠️ 舊資料歸屬修補',
      message: '這會把所有 creator 為空的舊資料，補成目前登入者。已經有歸屬的資料不會被覆蓋，是否繼續？',
      confirmLabel: '開始修補',
      confirmClass: 'btn btn-warning',
      onConfirm: async () => {
        const r = await fetch(API + '/api/creator-backfill', { method: 'POST' });
        const data = await r.json();
        if (!r.ok) return toast('修補失敗：' + (data.error || '未知錯誤'), 'error');
        toast(`修補完成：共 ${data.total_updated || 0} 筆`, 'success');
        const summary = document.getElementById('creator-backfill-summary');
        if (summary) {
          summary.innerHTML = `特殊權限帳號 ${data.privileged_accounts || 0} 筆 ｜ 平台申請 ${data.system_platform_requests || 0} 筆 ｜ 防火牆申請 ${data.firewall_requests || 0} 筆 ｜ 資產清冊 ${data.asset_inventory_records || 0} 筆 ｜ 防護基準 ${data.protection_baseline_records || 0} 筆`;
        }
        await Promise.all([loadOperationLogs(), loadDashboardPage()]);
      }
    });
  };

  window.loadOperationLogs = async function loadOperationLogs() {
    const r = await fetch(API + '/api/operation-logs?limit=200');
    const logs = await r.json();
    const tbody = document.getElementById('operation-logs-body');
    if (!tbody) return;
    if (!logs.length) {
      tbody.innerHTML = `<tr><td colspan="8"><div class="empty"><div class="icon">🧾</div><p>尚無操作日誌</p></div></td></tr>`;
      return;
    }
    tbody.innerHTML = logs.map(log => `
      <tr>
        <td>${log.id}</td>
        <td><span class="mono">${esc(log.event_type || '—')}</span></td>
        <td>${esc(formatDateTime(log.occurred_at))}</td>
        <td><span class="mono">${esc(log.location || log.request_path || '—')}</span></td>
        <td>${esc(log.user_name || '—')}<br><span class="hint">${esc(log.user_email || '未提供')}</span></td>
        <td><span class="hint">${esc(log.department || '—')}</span><br><span class="hint mono">${esc(log.user_google_id || '—')}</span></td>
        <td><span class="mono">${esc(log.source_ip || '—')}</span><br><span class="hint">${esc(log.request_method || '—')} / ${esc(String(log.status_code || '—'))}</span></td>
        <td class="hint">${esc(log.user_agent || '—')}</td>
      </tr>`).join('');
  };
})();
