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
    tr.innerHTML = '<td colspan="5" style="text-align:center;color:#6c757d">目前沒有零件資料</td>';
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
    items.push({ partId, partLabel, amount, note });
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
  setLoading(btn, true);
  try {
    for (const { partId, partLabel, amount, note } of items) {
      const { ok, data } = await apiFetch(`/parts/${partId}/decrease`, {
        method: 'POST',
        headers: { 'Idempotency-Key': generateIdempotencyKey() },
        body: JSON.stringify({ amount, note }),
      });
      if (!ok) {
        const reason = data?.error === 'insufficient stock' ? '庫存不足' : '操作失敗';
        showMessage('decrease-message', 'error', `${partLabel}：${reason}`);
        return;
      }
    }
    showMessage('decrease-message', 'success', `消耗紀錄已儲存，共 ${items.length} 筆`);
    loadStockPage('decrease');
  } finally {
    setLoading(btn, false);
  }
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

// ── Init ──────────────────────────────────────────────────────────────────────

showPage('parts');
