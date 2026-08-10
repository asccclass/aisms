(function () {
  async function ensurePartialLoaded(rootId, modalId, partialPath, errorMessage) {
    const root = document.getElementById(rootId);
    if (!root) return false;
    if (document.getElementById(modalId)) return true;

    const response = await fetch(partialPath);
    if (!response.ok) {
      toast(errorMessage, 'error');
      return false;
    }
    root.innerHTML = await response.text();
    return true;
  }

  window.ensureAccountModalLoaded = function ensureAccountModalLoaded() {
    return ensurePartialLoaded(
      'feature-account-modal-root',
      'account-modal',
      '/partials/account-modal.html',
      '帳號編輯視窗載入失敗'
    );
  };

  window.ensureNotifyModalLoaded = function ensureNotifyModalLoaded() {
    return ensurePartialLoaded(
      'feature-notify-modal-root',
      'notify-modal',
      '/partials/notify-modal.html',
      '通知視窗載入失敗'
    );
  };
})();
