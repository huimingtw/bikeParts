// ── Navigation ──────────────────────────────────────────────────────────────

const navLinks = document.querySelectorAll('.nav-link');
const pages = document.querySelectorAll('.page');

function showPage(pageId) {
  pages.forEach(p => { p.classList.remove('active'); });
  navLinks.forEach(l => { l.classList.remove('active'); });

  const page = document.getElementById(`page-${pageId}`);
  const link = document.querySelector(`.nav-link[data-page="${pageId}"]`);
  if (page) page.classList.add('active');
  if (link) link.classList.add('active');

  if (pageId === 'parts') loadParts();
  if (pageId === 'increase') loadStockPage('increase');
  if (pageId === 'decrease') loadStockPage('decrease');
  if (pageId === 'notifications') loadNotifications();
  if (pageId === 'settings') loadSettings();
}

navLinks.forEach(link => {
  link.addEventListener('click', e => {
    e.preventDefault();
    showPage(link.dataset.page);
  });
});

// ── Utilities ────────────────────────────────────────────────────────────────

function showMessage(elementId, type, text) {
  const el = document.getElementById(elementId);
  el.className = `message ${type}`;
  el.textContent = text;
  el.classList.remove('hidden');
  if (type === 'success') {
    setTimeout(() => el.classList.add('hidden'), 4000);
  }
}

function generateIdempotencyKey() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
    const r = Math.random() * 16 | 0;
    return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
  });
}

async function apiFetch(path, options = {}) {
  const res = await fetch(`/api${path}`, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  });
  const text = await res.text();
  const data = text ? JSON.parse(text) : {};
  return { ok: res.ok, status: res.status, data };
}

// ── Parts list ───────────────────────────────────────────────────────────────

async function loadParts() {
  const { ok, data } = await apiFetch('/parts');
  if (!ok) return;

  const tbody = document.getElementById('parts-tbody');
  tbody.innerHTML = '';

  if (!data || data.length === 0) {
    const tr = document.createElement('tr');
    tr.innerHTML = '<td colspan="6" style="text-align:center;color:#6c757d">目前沒有零件資料</td>';
    tbody.appendChild(tr);
    return;
  }

  data.forEach(part => {
    const tr = document.createElement('tr');
    const isLow = part.stock <= part.reorder_level;
    if (isLow) tr.classList.add('low-stock');
    tr.innerHTML = `
      <td>${part.sku}</td>
      <td>${part.name}</td>
      <td>${part.location || ''}</td>
      <td>${part.stock}${isLow ? ' ⚠' : ''}</td>
      <td>${part.reorder_level}</td>
      <td><button class="btn btn-secondary" style="padding:8px 16px;font-size:0.9rem">編輯</button></td>
    `;
    tr.querySelector('button').addEventListener('click', e => {
      e.stopPropagation();
      openEditForm(part);
    });
    tr.addEventListener('click', () => openEditForm(part));
    tbody.appendChild(tr);
  });
}

// ── Part form (create / edit) ─────────────────────────────────────────────────

document.getElementById('btn-new-part').addEventListener('click', () => {
  openCreateForm();
});

document.getElementById('btn-back-parts').addEventListener('click', () => {
  showPage('parts');
});

document.getElementById('btn-delete-part').addEventListener('click', async () => {
  const id = document.getElementById('part-id').value;
  if (!confirm('確定要刪除此零件嗎？')) return;

  const { ok } = await apiFetch(`/parts/${id}`, { method: 'DELETE' });
  if (ok) {
    showPage('parts');
  } else {
    showMessage('part-form-message', 'error', '刪除失敗，請再試一次');
  }
});

function openCreateForm() {
  document.getElementById('part-form-title').textContent = '新增零件';
  document.getElementById('part-id').value = '';
  document.getElementById('part-sku').value = '';
  document.getElementById('part-name').value = '';
  document.getElementById('part-location').value = '';
  document.getElementById('part-stock').value = '0';
  document.getElementById('part-reorder').value = '0';
  document.getElementById('part-sku').removeAttribute('readonly');
  document.getElementById('part-form-submit').textContent = '新增';
  document.getElementById('btn-delete-part').classList.add('hidden');
  document.getElementById('part-form-message').classList.add('hidden');

  pages.forEach(p => { p.classList.remove('active'); });
  navLinks.forEach(l => { l.classList.remove('active'); });
  document.getElementById('page-part-form').classList.add('active');
}

function openEditForm(part) {
  document.getElementById('part-form-title').textContent = '編輯零件';
  document.getElementById('part-id').value = part.id;
  document.getElementById('part-sku').value = part.sku;
  document.getElementById('part-name').value = part.name;
  document.getElementById('part-location').value = part.location || '';
  document.getElementById('part-stock').value = part.stock;
  document.getElementById('part-reorder').value = part.reorder_level;
  document.getElementById('part-sku').setAttribute('readonly', true);
  document.getElementById('part-form-submit').textContent = '儲存';
  document.getElementById('btn-delete-part').classList.remove('hidden');
  document.getElementById('part-form-message').classList.add('hidden');

  pages.forEach(p => { p.classList.remove('active'); });
  navLinks.forEach(l => { l.classList.remove('active'); });
  document.getElementById('page-part-form').classList.add('active');
}

document.getElementById('part-form').addEventListener('submit', async e => {
  e.preventDefault();
  const id = document.getElementById('part-id').value;
  const body = {
    sku:           document.getElementById('part-sku').value.trim(),
    name:          document.getElementById('part-name').value.trim(),
    location:      document.getElementById('part-location').value.trim(),
    stock:         parseInt(document.getElementById('part-stock').value, 10),
    reorder_level: parseInt(document.getElementById('part-reorder').value, 10),
  };

  const isEdit = !!id;
  const { ok } = isEdit
    ? await apiFetch(`/parts/${id}`, { method: 'PUT', body: JSON.stringify(body) })
    : await apiFetch('/parts', { method: 'POST', body: JSON.stringify(body) });

  if (ok) {
    showPage('parts');
  } else {
    showMessage('part-form-message', 'error', isEdit ? '更新失敗，請檢查輸入內容' : '新增失敗，請檢查輸入內容');
  }
});

// ── Stock list (increase / decrease) ─────────────────────────────────────────

let cachedParts = [];

async function loadStockPage(type) {
  const { ok, data } = await apiFetch('/parts');
  if (!ok) return;
  cachedParts = data || [];

  const listEl = document.getElementById(`${type}-list`);
  listEl.innerHTML = '';
  addStockRow(type); // start with one empty row
}

function buildPartOptions(selectedId) {
  const placeholder = '<option value="">-- 請選擇 --</option>';
  const opts = cachedParts.map(p =>
    `<option value="${p.id}" ${String(p.id) === String(selectedId) ? 'selected' : ''}>
      ${p.name}（${p.sku}）— 庫存：${p.stock}
    </option>`
  ).join('');
  return placeholder + opts;
}

function addStockRow(type, selectedId = '') {
  const listEl = document.getElementById(`${type}-list`);
  const row = document.createElement('div');
  row.className = 'stock-row';
  row.innerHTML = `
    <select class="row-part">${buildPartOptions(selectedId)}</select>
    <input type="number" class="row-amount" min="1" placeholder="數量" />
    <input type="text"   class="row-note"   placeholder="備註" />
    <button type="button" class="btn-remove" title="移除">✕</button>
  `;
  row.querySelector('.btn-remove').addEventListener('click', () => {
    row.remove();
  });
  listEl.appendChild(row);
}

function collectStockRows(type) {
  const rows = document.querySelectorAll(`#${type}-list .stock-row`);
  const items = [];
  for (const row of rows) {
    const select = row.querySelector('.row-part');
    const partId = select.value;
    const partLabel = select.options[select.selectedIndex]?.text || `零件 ${partId}`;
    const amount = parseInt(row.querySelector('.row-amount').value, 10);
    const note   = row.querySelector('.row-note').value.trim();
    if (!partId || !amount || amount < 1) return null; // invalid
    items.push({ partId, partLabel, amount, note, el: row });
  }
  return items;
}

document.getElementById('btn-add-increase').addEventListener('click', () => addStockRow('increase'));
document.getElementById('btn-add-decrease').addEventListener('click', () => addStockRow('decrease'));

function setLoading(btn, loading) {
  if (loading) {
    btn.dataset.originalText = btn.textContent;
    btn.textContent = '處理中…';
    btn.disabled = true;
  } else {
    btn.textContent = btn.dataset.originalText;
    btn.disabled = false;
  }
}

document.getElementById('btn-submit-increase').addEventListener('click', async (e) => {
  const btn = e.currentTarget;
  const items = collectStockRows('increase');
  if (!items || items.length === 0) {
    showMessage('increase-message', 'error', '請至少填寫一筆完整資料');
    return;
  }
  setLoading(btn, true);
  try {
    for (const { partId, amount, note } of items) {
      const { ok } = await apiFetch(`/parts/${partId}/increase`, {
        method: 'POST',
        headers: { 'Idempotency-Key': generateIdempotencyKey() },
        body: JSON.stringify({ amount, note }),
      });
      if (!ok) {
        showMessage('increase-message', 'error', '部分補貨失敗，請重新確認');
        return;
      }
    }
    showMessage('increase-message', 'success', `補貨成功，共 ${items.length} 筆`);
    loadStockPage('increase');
  } finally {
    setLoading(btn, false);
  }
});

document.getElementById('btn-submit-decrease').addEventListener('click', async (e) => {
  const btn = e.currentTarget;
  const items = collectStockRows('decrease');
  if (!items || items.length === 0) {
    showMessage('decrease-message', 'error', '請至少填寫一筆完整資料');
    return;
  }
  // Clear previous inline errors
  document.querySelectorAll('#decrease-list .row-error').forEach(el => el.remove());
  document.querySelectorAll('#decrease-list .stock-row').forEach(el => el.classList.remove('row-has-error'));

  setLoading(btn, true);
  const lowStockAlerts = [];
  let saved = 0;
  let failed = 0;
  try {
    for (const { partId, amount, note, el } of items) {
      const { ok, data } = await apiFetch(`/parts/${partId}/decrease`, {
        method: 'POST',
        headers: { 'Idempotency-Key': generateIdempotencyKey() },
        body: JSON.stringify({ amount, note }),
      });
      if (!ok) {
        const reason = data?.error === 'insufficient stock' ? '庫存不足' : '操作失敗';
        el.classList.add('row-has-error');
        const errEl = document.createElement('span');
        errEl.className = 'row-error';
        errEl.textContent = reason;
        el.appendChild(errEl);
        failed++;
        continue;
      }
      el.remove();
      saved++;
      if (data?.low_stock) {
        lowStockAlerts.push({ name: data.name, sku: data.sku, stock: data.stock, reorderLevel: data.reorder_level });
      }
    }
    if (failed > 0 && saved === 0) {
      showMessage('decrease-message', 'error', `${failed} 筆失敗，請修正後重新提交`);
    } else if (failed > 0) {
      showMessage('decrease-message', 'error', `已儲存 ${saved} 筆，${failed} 筆失敗，請修正後重新提交`);
    } else {
      showMessage('decrease-message', 'success', `消耗紀錄已儲存，共 ${saved} 筆`);
      loadStockPage('decrease');
    }
    if (lowStockAlerts.length > 0) showLowStockModal(lowStockAlerts);
  } finally {
    setLoading(btn, false);
  }
});

function showLowStockModal(alerts) {
  const list = document.getElementById('low-stock-modal-list');
  list.innerHTML = alerts.map(a =>
    `<li><strong>${a.name}</strong>（SKU: ${a.sku}）— 剩餘 <strong>${a.stock}</strong> 件 / 水位 ${a.reorderLevel} 件</li>`
  ).join('');
  document.getElementById('low-stock-modal').classList.remove('hidden');
}

document.getElementById('btn-close-low-stock-modal').addEventListener('click', () => {
  document.getElementById('low-stock-modal').classList.add('hidden');
});
document.getElementById('btn-confirm-low-stock-modal').addEventListener('click', () => {
  document.getElementById('low-stock-modal').classList.add('hidden');
});
document.getElementById('low-stock-modal').addEventListener('click', (e) => {
  if (e.target === e.currentTarget) e.currentTarget.classList.add('hidden');
});

// ── Notifications ─────────────────────────────────────────────────────────────

async function loadNotifications() {
  const { ok, data } = await apiFetch('/notifications');
  if (!ok) return;

  const tbody = document.getElementById('notifications-tbody');
  const empty = document.getElementById('notifications-empty');
  const table = document.getElementById('notifications-table');
  tbody.innerHTML = '';

  if (!data || data.length === 0) {
    table.classList.add('hidden');
    empty.classList.remove('hidden');
    return;
  }

  table.classList.remove('hidden');
  empty.classList.add('hidden');

  data.forEach(n => {
    const tr = document.createElement('tr');
    const date = new Date(n.created_at).toLocaleString('zh-TW');
    tr.innerHTML = `
      <td>${n.sku}</td>
      <td>${n.name}</td>
      <td>${n.reorder_level}</td>
      <td>${date}</td>
    `;
    tbody.appendChild(tr);
  });
}

document.getElementById('btn-refresh-notifications').addEventListener('click', loadNotifications);

// ── Settings ──────────────────────────────────────────────────────────────────

async function loadSettings() {
  const { ok, data } = await apiFetch('/settings');
  if (!ok) return;

  document.getElementById('settings-email-user').value    = data.email_user || '';
  document.getElementById('settings-email-pass').value    = data.email_pass || '';
  document.getElementById('settings-email-to').value      = data.email_to   || '';
  document.getElementById('settings-smtp-port').value     = data.smtp_port  || '587';
  document.getElementById('settings-port').value          = data.port       || '8080';
  document.getElementById('settings-email-enabled').checked = data.email_notifications_enabled ?? true;
}

document.getElementById('settings-form').addEventListener('submit', async e => {
  e.preventDefault();
  const body = {
    email_user:                  document.getElementById('settings-email-user').value.trim(),
    email_pass:                  document.getElementById('settings-email-pass').value,
    email_to:                    document.getElementById('settings-email-to').value.trim(),
    smtp_port:                   document.getElementById('settings-smtp-port').value.trim(),
    port:                        document.getElementById('settings-port').value.trim(),
    email_notifications_enabled: document.getElementById('settings-email-enabled').checked,
  };

  const { ok, data } = await apiFetch('/settings', { method: 'PUT', body: JSON.stringify(body) });
  if (ok) {
    const msg = data.restart_required
      ? '設定已儲存。Port 已變更，請重啟程式後生效。'
      : '設定已儲存。';
    showMessage('settings-message', 'success', msg);
    loadSettings(); // reload to show masked password
  } else {
    showMessage('settings-message', 'error', '儲存失敗，請再試一次');
  }
});

document.getElementById('btn-test-email').addEventListener('click', async () => {
  const btn = document.getElementById('btn-test-email');
  setLoading(btn, true);
  try {
    const { ok, data } = await apiFetch('/settings/test-email', { method: 'POST' });
    if (ok) {
      showMessage('settings-message', 'success', '測試信已寄出，請檢查收件匣。');
    } else {
      showMessage('settings-message', 'error', `寄送失敗：${data.error || '未知錯誤'}`);
    }
  } finally {
    setLoading(btn, false);
  }
});

// ── Setup Wizard ──────────────────────────────────────────────────────────────

async function checkFirstRun() {
  const { ok, data } = await apiFetch('/settings');
  if (ok && !data.is_configured) {
    document.getElementById('setup-overlay').classList.remove('hidden');
  }
}

document.getElementById('setup-form').addEventListener('submit', async e => {
  e.preventDefault();
  const body = {
    email_user: document.getElementById('setup-email-user').value.trim(),
    email_pass: document.getElementById('setup-email-pass').value,
    email_to:   document.getElementById('setup-email-to').value.trim(),
    smtp_port:  document.getElementById('setup-smtp-port').value.trim(),
    port:       '8080',
  };

  const { ok } = await apiFetch('/settings', { method: 'PUT', body: JSON.stringify(body) });
  if (ok) {
    document.getElementById('setup-overlay').classList.add('hidden');
    showMessage('parts-message', 'success', '設定完成，歡迎使用零件庫存系統！');
  } else {
    showMessage('setup-message', 'error', '儲存失敗，請再試一次');
  }
});

document.getElementById('btn-setup-test').addEventListener('click', async () => {
  const btn = document.getElementById('btn-setup-test');

  // Temporarily save settings before testing
  const body = {
    email_user: document.getElementById('setup-email-user').value.trim(),
    email_pass: document.getElementById('setup-email-pass').value,
    email_to:   document.getElementById('setup-email-to').value.trim(),
    smtp_port:  document.getElementById('setup-smtp-port').value.trim(),
    port:       '8080',
  };
  await apiFetch('/settings', { method: 'PUT', body: JSON.stringify(body) });

  setLoading(btn, true);
  try {
    const { ok, data } = await apiFetch('/settings/test-email', { method: 'POST' });
    if (ok) {
      showMessage('setup-message', 'success', '測試信已寄出，請檢查收件匣。');
    } else {
      showMessage('setup-message', 'error', `寄送失敗：${data.error || '未知錯誤'}`);
    }
  } finally {
    setLoading(btn, false);
  }
});

// ── Init ──────────────────────────────────────────────────────────────────────

checkFirstRun();
showPage('parts');
