// ===== Переключение вкладок =====
function switchTab(event, tabId) {
    document.querySelectorAll('.admin-tab').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    event.target.classList.add('active');
    document.getElementById(tabId).classList.add('active');

    if (tabId === 'bookings' && !bookingsLoaded) loadBookings();
}

// ===== Общие хелперы для работы с API =====
const API_BASE = '/api/v1';

async function apiFetch(path, options = {}) {
    const token = sessionStorage.getItem('adminToken');
    const headers = Object.assign(
        { 'Content-Type': 'application/json' },
        options.headers || {},
        token ? { 'Authorization': 'Bearer ' + token } : {}
    );

    const response = await fetch(API_BASE + path, Object.assign({}, options, { headers }));

    if (response.status === 401) {
        sessionStorage.clear();
        window.location.href = '/admin/';
        throw new Error('Сессия истекла');
    }

    const contentType = response.headers.get('content-type') || '';
    const isJson = contentType.includes('application/json');
    const data = isJson ? await response.json() : null;

    if (!response.ok) {
        throw new Error((data && data.error) || 'Ошибка запроса');
    }

    return data;
}

function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str === undefined || str === null ? '' : str;
    return div.innerHTML;
}

function formatDateTime(iso) {
    const d = new Date(iso);
    const datePart = d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'Asia/Yekaterinburg' });
    const timePart = d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', timeZone: 'Asia/Yekaterinburg' });
    return { datePart, timePart };
}

const STATUS_LABELS = {
    pending:   { text: 'Ожидает оплаты',   cls: 'status-pending' },
    confirmed: { text: 'Подтверждено',     cls: 'status-confirmed' },
    cancelled: { text: 'Отменено',         cls: 'status-cancelled' },
    completed: { text: 'Завершено',        cls: 'status-confirmed' },
    no_show:   { text: 'Клиент не пришёл', cls: 'status-cancelled' },
};

const PAYMENT_LABELS = {
    pending:  'оплата ожидается',
    paid:     'оплачено',
    failed:   'оплата не прошла',
    refunded: 'возврат оформлен',
};

// ===== ЗАПИСИ КЛИЕНТОВ =====
let bookingsLoaded = false;
let bookingsById = {};
let bookingActionBusy = false;

async function loadBookings() {
    const tbody = document.getElementById('bookingsTableBody');
    tbody.innerHTML = '<tr><td colspan="6">Загрузка…</td></tr>';

    try {
        const list = await apiFetch('/admin/appointments?limit=200');
        bookingsLoaded = true;
        renderBookings(list || []);
    } catch (err) {
        tbody.innerHTML = `<tr><td colspan="6">Не удалось загрузить записи: ${escapeHtml(err.message)}</td></tr>`;
    }
}

function renderBookings(list) {
    const tbody = document.getElementById('bookingsTableBody');
    bookingsById = {};
    list.forEach(a => { bookingsById[a.id] = a; });

    if (list.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6">Пока нет ни одной записи</td></tr>';
        return;
    }

    tbody.innerHTML = list.map(appt => {
        const client = appt.client || {};
        const service = appt.service || {};
        const { datePart, timePart } = formatDateTime(appt.starts_at);
        const statusInfo = STATUS_LABELS[appt.status] || { text: appt.status, cls: 'status-pending' };
        const paymentLabel = (PAYMENT_LABELS[appt.payment_status] || appt.payment_status) + paymentSuffix(appt);

        return `
            <tr data-id="${appt.id}">
                <td>
                    <strong>${escapeHtml(client.first_name)} ${escapeHtml(client.last_name)}</strong><br>
                    <span style="font-size: 12px; color: var(--text-muted);">${escapeHtml(client.phone)} · ${escapeHtml(client.email)}</span>
                </td>
                <td>${escapeHtml(service.title)}</td>
                <td>${datePart}<br>${timePart}</td>
                <td>${appt.format === 'online' ? 'Онлайн' : 'Очно'}</td>
                <td>
                    <span class="status ${statusInfo.cls}">${statusInfo.text}</span><br>
                    <span style="font-size: 11px; color: var(--text-muted);">${escapeHtml(paymentLabel)}</span>
                </td>
                <td>
                    ${bookingActionsHtml(appt)}
                </td>
            </tr>
        `;
    }).join('');
}

async function cancelBooking(id) {
    if (bookingActionBusy) return;
    const appt = bookingsById[id];
    if (!appt) { loadBookings(); return; }

    const isPaid = appt.payment_status === 'paid';
    let refund = false;

    if (isPaid && appt.payment_id) {
        const choice = await askRefundChoice(appt);
        if (!choice) return;
        refund = choice === 'refund';
    } else if (isPaid) {
        if (!confirm('Эта запись оплачена вручную (не через ЮKassa), поэтому вернуть деньги автоматически нельзя.\n\nОтменить запись? Деньги придётся вернуть клиенту самостоятельно.')) return;
    } else {
        if (!confirm('Отменить эту запись? Клиент не будет уведомлён автоматически.')) return;
    }

    bookingActionBusy = true;
    try {
        const res = await apiFetch(`/admin/appointments/${id}/cancel`, {
            method: 'POST',
            body: JSON.stringify({ refund }),
        });
        if (res && res.refunded) {
            alert(res.refund_status === 'succeeded'
                ? 'Запись отменена, деньги возвращены клиенту.'
                : 'Запись отменена. Возврат принят ЮKassa, деньги придут клиенту в течение нескольких рабочих дней.');
        } else if (isPaid) {
            alert('Запись отменена без возврата денег.');
        }
    } catch (err) {
        alert('Не удалось отменить запись: ' + err.message);
    } finally {
        bookingActionBusy = false;
        loadBookings();
    }
}

async function deleteBooking(id) {
    if (!confirm('Удалить запись безвозвратно? Это действие нельзя отменить.')) return;
    try {
        await apiFetch(`/admin/appointments/${id}`, { method: 'DELETE' });
        loadBookings();
    } catch (err) {
        alert('Не удалось удалить запись: ' + err.message);
    }
}

async function rescheduleBooking(id) {
    const input = prompt('Новая дата и время (в формате ГГГГ-ММ-ДД ЧЧ:ММ), например 2026-07-25 14:00:');
    if (!input) return;

    const match = input.trim().match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2})$/);
    if (!match) {
        alert('Неверный формат. Пример: 2026-07-25 14:00');
        return;
    }
    const [, y, mo, d, h, mi] = match;
    const startsAt = toCalendarISO(Number(y), Number(mo), Number(d), Number(h), Number(mi));

    try {
        await apiFetch(`/admin/appointments/${id}/reschedule`, {
            method: 'PATCH',
            body: JSON.stringify({ starts_at: startsAt }),
        });
        loadBookings();
    } catch (err) {
        alert('Не удалось перенести запись: ' + err.message);
    }
}


// ===== Ручное создание записи (клиент договорился по телефону и т.п.) =====
let servicesForSelect = [];

async function ensureServiceOptionsLoaded() {
    if (servicesForSelect.length > 0) return;
    try {
        const response = await fetch('/api/v1/services');
        servicesForSelect = await response.json() || [];
        const select = document.getElementById('manualServiceSelect');
        select.innerHTML = servicesForSelect
            .map(s => `<option value="${s.id}">${escapeHtml(s.title)} — ${(s.price_kopeks / 100).toLocaleString('ru-RU')} ₽</option>`)
            .join('');
    } catch (err) {
        console.error('Не удалось загрузить список услуг:', err);
    }
}

async function toggleManualBookingForm() {
    const form = document.getElementById('manualBookingForm');
    const isHidden = form.style.display === 'none' || !form.style.display;
    form.style.display = isHidden ? 'block' : 'none';
    if (isHidden) {
        await ensureServiceOptionsLoaded();
        document.getElementById('manualFirstName').focus();
    } else {
        clearManualBookingForm();
    }
}

function clearManualBookingForm() {
    ['manualFirstName', 'manualLastName', 'manualPatronym', 'manualPhone', 'manualEmail', 'manualDate', 'manualTime', 'manualNotes']
        .forEach(id => { document.getElementById(id).value = ''; });
}

async function saveManualBooking() {
    const firstName = document.getElementById('manualFirstName').value.trim();
    const lastName = document.getElementById('manualLastName').value.trim();
    const patronym = document.getElementById('manualPatronym').value.trim();
    const phone = document.getElementById('manualPhone').value.trim();
    const email = document.getElementById('manualEmail').value.trim();
    const serviceId = document.getElementById('manualServiceSelect').value;
    const dateStr = document.getElementById('manualDate').value;
    const timeStr = document.getElementById('manualTime').value;
    const format = document.getElementById('manualFormat').value;
    const paymentStatus = document.getElementById('manualPaymentStatus').value;
    const notes = document.getElementById('manualNotes').value.trim();

    if (!firstName || !lastName || !phone || !email || !serviceId || !dateStr || !timeStr) {
        alert('Заполните обязательные поля: имя, фамилия, телефон, email, услуга, дата и время');
        return;
    }

    const [y, mo, d] = dateStr.split('-').map(Number);
    const [hh, mi] = timeStr.split(':').map(Number);
    const startsAt = toCalendarISO(y, mo, d, hh, mi);

    try {
        await apiFetch('/admin/appointments', {
            method: 'POST',
            body: JSON.stringify({
                client: { first_name: firstName, last_name: lastName, patronym, phone, email },
                service_id: serviceId,
                format,
                starts_at: startsAt,
                payment_status: paymentStatus,
                notes,
            }),
        });
        clearManualBookingForm();
        toggleManualBookingForm();
        alert('Запись создана, клиенту отправлено письмо с подтверждением');
        loadBookings();
    } catch (err) {
        alert('Не удалось создать запись: ' + err.message);
    }
}

// Время записи привязано к часовому поясу кабинета (Asia/Yekaterinburg, +05:00),
// а не к часовому поясу браузера. Бэкенд принимает только RFC3339 со смещением +05:00.
function toCalendarISO(year, month1, day, hh, mm) {
    const p = n => String(n).padStart(2, '0');
    return `${year}-${p(month1)}-${p(day)}T${p(hh)}:${p(mm)}:00+05:00`;
}


// ===== Действия над записями: кнопки, возврат, статусы =====
function formatRubles(kopeks) {
    return (kopeks / 100).toLocaleString('ru-RU', { maximumFractionDigits: 2 }) + ' ₽';
}

function paymentSuffix(appt) {
    let suffix = '';
    if (appt.amount_kopeks > 0) suffix += ' · ' + formatRubles(appt.amount_kopeks);
    if (appt.payment_mode === 'prepay_50') suffix += ' · предоплата 50%';
    return suffix;
}

function bookingBtn(label, handler, danger) {
    return `<button class="btn btn-small${danger ? ' btn-danger' : ''}" onclick="${handler}">${label}</button>`;
}

function bookingActionsHtml(appt) {
    const id = appt.id;
    const isActive = appt.status === 'pending' || appt.status === 'confirmed';
    const isPaid = appt.payment_status === 'paid';
    const started = new Date(appt.starts_at).getTime() <= Date.now();
    const parts = [];

    if (isActive) parts.push(bookingBtn('Перенести', `rescheduleBooking('${id}')`));
    if (appt.status === 'confirmed' && appt.payment_status === 'pending' && !appt.payment_id) {
        parts.push(bookingBtn('Отметить оплаченной', `markBookingPaid('${id}')`));
    }
    if (appt.status === 'confirmed' && started) {
        parts.push(bookingBtn('Завершено', `setBookingStatus('${id}', 'completed')`));
        parts.push(bookingBtn('Не пришёл', `setBookingStatus('${id}', 'no_show')`, true));
    }
    if (isActive) parts.push(bookingBtn('Отменить', `cancelBooking('${id}')`, true));
    // Оплаченную запись удалять нельзя: сначала отмена (с возвратом или без)
    if (!isPaid) parts.push(bookingBtn('Удалить', `deleteBooking('${id}')`, true));

    return parts.join(' ');
}

// Окно «Вернуть клиенту деньги?». Возвращает 'refund', 'keep' или null (не отменять).
function askRefundChoice(appt) {
    return new Promise(resolve => {
        const client = appt.client || {};
        const service = appt.service || {};
        const { datePart, timePart } = formatDateTime(appt.starts_at);

        const overlay = document.createElement('div');
        overlay.style.cssText = 'position:fixed;inset:0;background:rgba(0,0,0,.45);display:flex;align-items:center;justify-content:center;z-index:10000;padding:16px;';

        const card = document.createElement('div');
        card.style.cssText = 'background:#fff;border-radius:12px;max-width:440px;width:100%;padding:24px;box-shadow:0 10px 40px rgba(0,0,0,.25);';
        card.innerHTML = `
            <div style="font-size:20px;font-weight:600;margin-bottom:10px;">Вернуть клиенту деньги?</div>
            <div style="font-size:14px;line-height:1.6;margin-bottom:12px;">
                <strong>${escapeHtml(client.first_name)} ${escapeHtml(client.last_name)}</strong><br>
                ${escapeHtml(service.title)}<br>
                ${datePart}, ${timePart}
                ${appt.amount_kopeks > 0 ? `<br>Оплачено: ${formatRubles(appt.amount_kopeks)}` : ''}
            </div>
            <div style="font-size:13px;color:#666;line-height:1.5;margin-bottom:16px;">
                «Да» — запись отменится, деньги вернутся клиенту через ЮKassa, клиент получит письмо.<br>
                «Нет» — запись отменится без возврата денег.
            </div>
            <div style="display:flex;flex-direction:column;gap:8px;">
                <button type="button" class="btn" data-choice="refund">Да, вернуть деньги</button>
                <button type="button" class="btn btn-danger" data-choice="keep">Нет, отменить без возврата</button>
                <button type="button" class="btn" data-choice="back" style="background:#eee;color:#333;">Не отменять</button>
            </div>
        `;

        function close(choice) {
            document.removeEventListener('keydown', onKey);
            overlay.remove();
            resolve(choice);
        }
        function onKey(e) { if (e.key === 'Escape') close(null); }

        card.querySelectorAll('button').forEach(btn => {
            btn.addEventListener('click', () => close(btn.dataset.choice === 'back' ? null : btn.dataset.choice));
        });
        overlay.addEventListener('click', e => { if (e.target === overlay) close(null); });
        document.addEventListener('keydown', onKey);

        overlay.appendChild(card);
        document.body.appendChild(overlay);
    });
}

async function setBookingStatus(id, status) {
    if (bookingActionBusy) return;
    const question = status === 'completed'
        ? 'Отметить приём как состоявшийся? Изменить это потом будет нельзя.'
        : 'Отметить, что клиент не пришёл? Деньги по записи не возвращаются, изменить это потом будет нельзя.';
    if (!confirm(question)) return;

    bookingActionBusy = true;
    try {
        await apiFetch(`/admin/appointments/${id}/status`, {
            method: 'PATCH',
            body: JSON.stringify({ status }),
        });
    } catch (err) {
        alert('Не удалось изменить статус: ' + err.message);
    } finally {
        bookingActionBusy = false;
        loadBookings();
    }
}

async function markBookingPaid(id) {
    if (bookingActionBusy) return;
    if (!confirm('Отметить запись как оплаченную (наличные или перевод)?')) return;

    bookingActionBusy = true;
    try {
        await apiFetch(`/admin/appointments/${id}/mark-paid`, { method: 'POST' });
    } catch (err) {
        alert('Не удалось отметить оплату: ' + err.message);
    } finally {
        bookingActionBusy = false;
        loadBookings();
    }
}

// Кнопка «Выйти» в шапке (функция logout() объявлена в admin.html)
document.addEventListener('DOMContentLoaded', () => {
    const link = document.querySelector('.admin-logout');
    if (link) link.addEventListener('click', e => { e.preventDefault(); logout(); });
});
