const API_URL_LIST = '/api/orders';
const API_URL_BY_UID = '/api/orders/uid/';
const API_URL_CREATE = '/api/fake/generate';

const searchTab = document.getElementById('searchTab');
const listTab = document.getElementById('listTab');
const searchSection = document.getElementById('searchSection');
const listSection = document.getElementById('listSection');

const uidForm = document.getElementById('uidForm');
const uidInput = document.getElementById('uidInput');
const uidError = document.getElementById('uidError');
const orderDetail = document.getElementById('orderDetail');

const loadAllBtn = document.getElementById('loadAllBtn');
const createOrderBtn = document.getElementById('createOrderBtn');
const clearBtn = document.getElementById('clearBtn');
const autoRefreshCheckbox = document.getElementById('autoRefresh');
const orderList = document.getElementById('orderList');
const orderDetailList = document.getElementById('orderDetailList');

let ordersCache = [];
let selectedUid = null;
let autoRefreshTimer = null;
const AUTO_REFRESH_INTERVAL = 30_000;

searchTab.addEventListener('click', () => switchTab('search'));
listTab.addEventListener('click', () => switchTab('list'));

function switchTab(tabName) {
  searchTab.classList.toggle('active', tabName === 'search');
  listTab.classList.toggle('active', tabName === 'list');
  
  searchSection.classList.toggle('active', tabName === 'search');
  listSection.classList.toggle('active', tabName === 'list');
  
  uidError.textContent = '';
}

uidForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  uidError.textContent = '';
  const uid = uidInput.value.trim();
  if (!uid) {
    uidError.textContent = 'UID заказа обязателен';
    return;
  }
  await fetchByUID(uid);
});

loadAllBtn.addEventListener('click', loadAllOrders);
createOrderBtn.addEventListener('click', createOrder);
clearBtn.addEventListener('click', () => {
  ordersCache = [];
  renderList([]);
  renderDetailList(null);
});
autoRefreshCheckbox.addEventListener('change', (e) => {
  if (e.target.checked) {
    startAutoRefresh();
  } else {
    stopAutoRefresh();
  }
});

async function fetchByUID(uid) {
  renderDetailLoading(orderDetail);
  try {
    const res = await fetch(API_URL_BY_UID + encodeURIComponent(uid), {
      headers: { 'Accept': 'application/json' }
    });
    
    if (res.ok) {
      const order = await res.json();
      renderDetail(order, orderDetail);
    } else {
      let body = '';
      try {
        body = await res.json();
        body = JSON.stringify(body);
      } catch (e) {
        try {
          body = await res.text();
        } catch (e2) {
          body = '';
        }
      }
      renderDetailError(`Ошибка: ${res.status} ${res.statusText} ${body ? '- ' + body : ''}`, orderDetail);
    }
  } catch (err) {
    renderDetailError('Сетевая ошибка: ' + String(err), orderDetail);
  }
}

async function createOrder() {
  const originalText = createOrderBtn.textContent;
  createOrderBtn.textContent = 'Создание...';
  createOrderBtn.disabled = true;
  
  try {
    const res = await fetch(API_URL_CREATE, {
      method: 'POST',
      headers: { 'Accept': 'application/json' }
    });
    
    if (res.ok) {
      const result = await res.json();
      
      showNotification(`Заказ успешно создан! UID: ${result.order_uid}`, 'success');
      
      if (ordersCache.length > 0) {
        await loadAllOrders();
      }
      
      if (result.order) {
        renderDetailList(result.order);
        selectedUid = result.order.order_uid;
        highlightSelectedInList(selectedUid);
      }
    } else {
      let errorMsg = 'Ошибка при создании заказа';
      try {
        const errorData = await res.json();
        errorMsg += ': ' + (errorData.error || res.statusText);
      } catch (e) {
        errorMsg += ': ' + res.statusText;
      }
      showNotification(errorMsg, 'error');
    }
  } catch (err) {
    showNotification('Сетевая ошибка при создании заказа: ' + String(err), 'error');
  } finally {
    createOrderBtn.textContent = originalText;
    createOrderBtn.disabled = false;
  }
}

async function loadAllOrders() {
  orderList.innerHTML = `
    <div style="padding:18px">
      <div class="spinner" aria-hidden="true"></div>
      <div style="text-align:center;color:var(--muted);font-size:14px;margin-top:8px">
        Загрузка списка заказов...
      </div>
    </div>
  `;
  
  try {
    const res = await fetch(API_URL_LIST, {
      headers: { 'Accept': 'application/json' }
    });
    
    if (!res.ok) throw new Error('HTTP ' + res.status);
    
    const data = await res.json();
    if (!Array.isArray(data)) {
      throw new Error('Ожидался массив заказов, пришло: ' + typeof data);
    }
    
    ordersCache = data;
    renderList(ordersCache);
    
    const sel = ordersCache.find(o => o.order_uid === selectedUid);
    if (sel) {
      renderDetailList(sel);
    }
  } catch (err) {
    orderList.innerHTML = `
      <div class="empty">
        Ошибка при загрузке списка: ${escapeHtml(String(err))}
      </div>
    `;
    console.error(err);
  }
}

function startAutoRefresh() {
  stopAutoRefresh();
  autoRefreshTimer = setInterval(loadAllOrders, AUTO_REFRESH_INTERVAL);
  loadAllBtn.disabled = true;
}

function stopAutoRefresh() {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer);
    autoRefreshTimer = null;
  }
  loadAllBtn.disabled = false;
}

function renderList(list) {
  if (!list || list.length === 0) {
    orderList.innerHTML = `
      <div class="empty">
        Список пуст — нажмите «Загрузить все заказы» для отображения списка
      </div>
    `;
    return;
  }
  
  const frag = document.createDocumentFragment();
  list.forEach(order => {
    const row = document.createElement('div');
    row.className = 'order-row';
    if (order.order_uid === selectedUid) {
      row.classList.add('selected');
    }
    row.tabIndex = 0;

    const left = document.createElement('div');
    left.style.display = 'flex';
    left.style.flexDirection = 'column';

    const uid = document.createElement('div');
    uid.className = 'order-uid';
    uid.textContent = order.order_uid;

    const meta = document.createElement('div');
    meta.className = 'row-meta';
    const date = order.date_created ? formatDate(order.date_created) : '';
    const items = (order.items && order.items.length) ? `${order.items.length} шт` : '0';
    meta.textContent = `${date} • ${items}`;

    left.appendChild(uid);
    left.appendChild(meta);
    row.appendChild(left);

    row.addEventListener('click', () => selectAndShow(order));
    row.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        selectAndShow(order);
      }
    });

    frag.appendChild(row);
  });
  
  orderList.innerHTML = '';
  orderList.appendChild(frag);
}

function selectAndShow(order) {
  selectedUid = order.order_uid;
  highlightSelectedInList(selectedUid);
  renderDetailList(order);
}

function highlightSelectedInList(uid) {
  [...orderList.querySelectorAll('.order-row')].forEach(r => {
    const text = r.querySelector('.order-uid')?.textContent || '';
    r.classList.toggle('selected', text === uid);
  });
}

function renderDetailLoading(element) {
  element.innerHTML = `
    <div class="placeholder">
      <div class="spinner" aria-hidden="true"></div>
      <div style="text-align:center;color:var(--muted);font-size:14px;margin-top:8px">
        Загрузка заказа...
      </div>
    </div>
  `;
}

function renderDetailError(msg, element) {
  element.innerHTML = `
    <div class="placeholder" style="color:var(--danger)">
      ${escapeHtml(msg)}
    </div>
  `;
}

function renderDetail(order, element) {
  if (!order) {
    element.innerHTML = `
      <div class="placeholder">
        Введите UID заказа для поиска
      </div>
    `;
    return;
  }

  const d = order.delivery || {};
  const p = order.payment || {};
  const items = order.items || [];

  const html = `
    <div class="card">
      <h3>Общая информация</h3>
      <div class="kv"><b>UID заказа</b><span>${escapeHtml(order.order_uid)}</span></div>
      <div class="kv"><b>Трек номер</b><span>${escapeHtml(order.track_number || '')}</span></div>
      <div class="kv"><b>Дата создания</b><span>${escapeHtml(formatDate(order.date_created))}</span></div>
    </div>

    <div class="card">
      <h3>Информация о доставке</h3>
      <div class="kv"><b>Имя получателя</b><span>${escapeHtml(d.name || '')}</span></div>
      <div class="kv"><b>Телефон</b><span>${escapeHtml(d.phone || '')}</span></div>
      <div class="kv"><b>Адрес</b><span>${escapeHtml([d.city, d.address].filter(Boolean).join(', '))}</span></div>
      <div class="kv"><b>Email</b><span>${escapeHtml(d.email || '')}</span></div>
    </div>

    <div class="card">
      <h3>Информация о платеже</h3>
      <div class="kv"><b>Сумма</b><span>${escapeHtml((p.amount != null) ? p.amount + ' ' + (p.currency || '') : '')}</span></div>
      <div class="kv"><b>Провайдер</b><span>${escapeHtml(p.provider || '')}</span></div>
      <div class="kv"><b>Дата платежа</b><span>${escapeHtml(formatUnixOrIso(p.payment_dt))}</span></div>
      <div class="kv"><b>Банк</b><span>${escapeHtml(p.bank || '')}</span></div>
    </div>

    <div class="card">
      <h3>Товары (${items.length})</h3>
      ${items.length ? renderItemsTable(items) : '<div style="color:var(--muted);font-size:14px">Нет товаров</div>'}
    </div>

    <div class="card">
      <h3>Исходные данные JSON</h3>
      <div style="display:flex;gap:8px;align-items:center;margin-bottom:8px">
        <button id="copyJsonBtn" class="btn primary">Копировать JSON</button>
        <button id="toggleRawBtn" class="btn secondary">Показать/скрыть</button>
      </div>
      <pre id="rawJson" class="json" style="display:none">${escapeHtml(JSON.stringify(order, null, 2))}</pre>
    </div>
  `;
  
  element.innerHTML = html;

  const copyBtn = element.querySelector('#copyJsonBtn');
  const toggleBtn = element.querySelector('#toggleRawBtn');
  
  if (copyBtn) {
    copyBtn.addEventListener('click', async () => {
      try {
        await navigator.clipboard.writeText(JSON.stringify(order, null, 2));
        copyBtn.textContent = 'Скопировано!';
        setTimeout(() => copyBtn.textContent = 'Копировать JSON', 1200);
      } catch (e) {
        alert('Не удалось скопировать: ' + e);
      }
    });
  }
  
  if (toggleBtn) {
    toggleBtn.addEventListener('click', () => {
      const raw = element.querySelector('#rawJson');
      raw.style.display = raw.style.display === 'none' ? 'block' : 'none';
    });
  }
}

function renderDetailList(order) {
  renderDetail(order, orderDetailList);
}

function renderItemsTable(items) {
  const rows = items.map(it => {
    return `<tr>
      <td>${escapeHtml(it.nm_id)}</td>
      <td>${escapeHtml(it.name)}</td>
      <td>${escapeHtml(it.brand || '')}</td>
      <td>${escapeHtml(String(it.price || ''))}</td>
      <td>${escapeHtml(String(it.sale || ''))}</td>
      <td>${escapeHtml(String(it.total_price || ''))}</td>
    </tr>`;
  }).join('');
  
  return `<table class="items">
    <thead>
      <tr>
        <th>nm_id</th>
        <th>Название</th>
        <th>Бренд</th>
        <th>Цена</th>
        <th>Скидка</th>
        <th>Итого</th>
      </tr>
    </thead>
    <tbody>${rows}</tbody>
  </table>`;
}

function showNotification(message, type = 'info') {
  const notification = document.createElement('div');
  notification.className = `notification notification-${type}`;
  notification.innerHTML = `
    <div class="notification-content">
      <span class="notification-message">${escapeHtml(message)}</span>
      <button class="notification-close" onclick="this.parentElement.parentElement.remove()">×</button>
    </div>
  `;

  document.body.insertBefore(notification, document.body.firstChild);

  setTimeout(() => {
    if (notification.parentElement) {
      notification.remove();
    }
  }, 5000);
}

function escapeHtml(str) {
  if (str == null) return '';
  return String(str).replace(/[&<>"']/g, s => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  })[s]);
}

function formatDate(s) {
  if (!s) return '';
  try {
    if (typeof s === 'number') {
      return new Date(s * 1000).toLocaleString();
    }
    const d = new Date(s);
    if (!isNaN(d)) {
      return d.toLocaleString();
    }
    return String(s);
  } catch (e) {
    return String(s);
  }
}

function formatUnixOrIso(v) {
  if (!v) return '';
  return (typeof v === 'number') 
    ? new Date(v * 1000).toLocaleString() 
    : new Date(v).toLocaleString();
}
