(function () {
  const modalRootId = 'feature-export-docx-root';
  const modalPartialPath = '/partials/export-docx-modal.html';
  const storagePrefix = 'isms_export_docx_signers';

  function getCurrentUserName() {
    return document.getElementById('user-profile-acc')?.dataset?.userName || '';
  }

  function buildStorageKey() {
    const user = getCurrentUserName().trim().toLowerCase();
    return `${storagePrefix}:${user || 'anonymous'}`;
  }

  function readSavedSigners() {
    try {
      const raw = window.localStorage.getItem(buildStorageKey());
      return raw ? JSON.parse(raw) : {};
    } catch (_) {
      return {};
    }
  }

  function saveSigners(inventoryBy, groupLeader) {
    try {
      window.localStorage.setItem(buildStorageKey(), JSON.stringify({
        inventory_by: inventoryBy,
        group_leader: groupLeader
      }));
    } catch (_) {
      // Ignore storage failures and continue export flow.
    }
  }

  async function ensureModalLoaded() {
    const root = document.getElementById(modalRootId);
    if (!root) return false;
    if (document.getElementById('export-docx-modal')) return true;

    const response = await fetch(modalPartialPath);
    if (!response.ok) {
      toast('DOCX 匯出視窗載入失敗', 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  window.downloadAccountsDocx = async function downloadAccountsDocx() {
    const loaded = await ensureModalLoaded();
    if (!loaded) return;

    const saved = readSavedSigners();
    const currentUser = getCurrentUserName();
    document.getElementById('export-inventory-by').value = saved.inventory_by || currentUser;
    document.getElementById('export-group-leader').value = saved.group_leader || '';
    document.getElementById('export-docx-modal').classList.add('open');
  };

  window.confirmDownloadAccountsDocx = function confirmDownloadAccountsDocx() {
    const status = document.getElementById('status-filter').value;
    const q = document.getElementById('search-input').value.trim();
    const params = buildAccountQueryParams(status, q);
    const inventoryBy = document.getElementById('export-inventory-by').value.trim();
    const groupLeader = document.getElementById('export-group-leader').value.trim();

    if (inventoryBy) params.set('inventory_by', inventoryBy);
    if (groupLeader) params.set('group_leader', groupLeader);

    saveSigners(inventoryBy, groupLeader);

    let url = API + '/api/accounts/export-docx';
    if ([...params].length) url += '?' + params.toString();
    closeModal('export-docx-modal');
    window.location.href = url;
  };
})();
