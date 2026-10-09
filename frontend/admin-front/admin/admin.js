// ===== Переключение вкладок =====
function switchTab(event, tabId) {
    document.querySelectorAll('.admin-tab').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    event.target.classList.add('active');
    document.getElementById(tabId).classList.add('active');

    if (tabId === 'bookings' && !bookingsLoaded) loadBookings();
    if (tabId === 'schedule' && !scheduleLoaded) initSchedule();
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
        const error = new Error((data && data.error) || 'Ошибка запроса');
        error.status = response.status;
        error.data = data;
        throw error;
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


// ===== РАСПИСАНИЕ =====
const SCHED_SLOTS = ['10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00'];
const SCHED_MONTHS = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь', 'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'];
const SCHED_MONTHS_GEN = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня', 'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря'];
const SCHED_WEEKDAYS = ['воскресенье', 'понедельник', 'вторник', 'среда', 'четверг', 'пятница', 'суббота'];
const SCHED_DAY_NAMES = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

let scheduleLoaded = false;
let schedBusy = false;
let schedYear = 0;
let schedMonth = 0;
let schedSelected = null;
let schedData = { slots: SCHED_SLOTS, blocks: [], appointments: [] };

function schedPad(n) { return String(n).padStart(2, '0'); }
function schedKey(y, m0, d) { return `${y}-${schedPad(m0 + 1)}-${schedPad(d)}`; }

function schedTodayKey() {
    const parts = new Intl.DateTimeFormat('en-US', {
        timeZone: 'Asia/Yekaterinburg', year: 'numeric', month: '2-digit', day: '2-digit',
    }).formatToParts(new Date());
    const get = type => parts.find(p => p.type === type).value;
    return `${get('year')}-${get('month')}-${get('day')}`;
}

function schedPlural(n) {
    const m10 = n % 10, m100 = n % 100;
    if (m10 === 1 && m100 !== 11) return 'запись';
    if (m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14)) return 'записи';
    return 'записей';
}

function schedEntry(idx, key) {
    return idx[key] || { fullBlock: null, slotBlocks: {}, booked: {}, appts: [] };
}

function schedIndex() {
    const idx = {};
    const ensure = date => idx[date] || (idx[date] = { fullBlock: null, slotBlocks: {}, booked: {}, appts: [] });
    schedData.blocks.forEach(b => {
        const e = ensure(b.date);
        if (b.slot_time) e.slotBlocks[b.slot_time] = b; else e.fullBlock = b;
    });
    schedData.appointments.forEach(a => {
        const e = ensure(a.date);
        e.appts.push(a);
        (a.slots || []).forEach(t => { e.booked[t] = a; });
    });
    return idx;
}

function schedDayInfo(e, slots) {
    const blocked = new Set(e.fullBlock ? slots : Object.keys(e.slotBlocks));
    const busy = new Set([...blocked, ...Object.keys(e.booked)]);
    return {
        closed: !!e.fullBlock,
        blockedCount: blocked.size,
        full: busy.size >= slots.length,
        apptCount: e.appts.length,
    };
}

function initSchedule() {
    const [y, m] = schedTodayKey().split('-').map(Number);
    schedYear = y;
    schedMonth = m - 1;
    schedSelected = schedTodayKey();

    const hourFrom = document.getElementById('schedHourFrom');
    const hourTo = document.getElementById('schedHourTo');
    hourFrom.innerHTML = SCHED_SLOTS.map(t => `<option value="${t}">${t}</option>`).join('');
    hourTo.innerHTML = SCHED_SLOTS.map(t => {
        const end = schedPad(Number(t.slice(0, 2)) + 1) + ':00';
        return `<option value="${end}">${end}</option>`;
    }).join('');
    hourTo.value = '20:00';
    document.getElementById('schedFrom').min = schedTodayKey();
    document.getElementById('schedTo').min = schedTodayKey();

    loadSchedule();
}

async function loadSchedule() {
    const grid = document.getElementById('schedCalendar');
    try {
        const data = await apiFetch(`/admin/schedule?month=${schedYear}-${schedPad(schedMonth + 1)}`);
        schedData = {
            slots: (data.slots && data.slots.length) ? data.slots : SCHED_SLOTS,
            blocks: data.blocks || [],
            appointments: data.appointments || [],
        };
        scheduleLoaded = true;
        renderSchedule();
    } catch (err) {
        grid.innerHTML = `<div class="sch-error">Не удалось загрузить расписание: ${escapeHtml(err.message)}</div>`;
    }
}

function renderSchedule() {
    const slots = schedData.slots;
    const idx = schedIndex();
    const today = schedTodayKey();

    document.getElementById('schedMonthTitle').textContent = SCHED_MONTHS[schedMonth] + ' ' + schedYear;

    const offset = (new Date(schedYear, schedMonth, 1).getDay() + 6) % 7;
    const daysInMonth = new Date(schedYear, schedMonth + 1, 0).getDate();

    let html = SCHED_DAY_NAMES.map(n => `<div class="calendar-day-name">${n}</div>`).join('');
    for (let i = 0; i < offset; i++) html += '<div class="calendar-day empty"></div>';

    for (let d = 1; d <= daysInMonth; d++) {
        const key = schedKey(schedYear, schedMonth, d);
        const info = schedDayInfo(schedEntry(idx, key), slots);
        const cls = ['calendar-day', 'sch-day'];
        if (key < today) cls.push('sch-past');
        if (key === today) cls.push('sch-today');
        if (key === schedSelected) cls.push('selected');
        if (info.closed) cls.push('sch-closed');
        else if (info.full) cls.push('sch-full');
        else if (info.blockedCount > 0) cls.push('sch-partial');
        if (info.apptCount > 0) cls.push('sch-has');

        const notes = [];
        if (info.closed) notes.push('закрыт');
        else if (info.blockedCount > 0) notes.push(`закрыто ${info.blockedCount} ч`);
        if (info.apptCount > 0) notes.push(`${info.apptCount} ${schedPlural(info.apptCount)}`);

        html += `<div class="${cls.join(' ')}" data-date="${key}" onclick="schedSelectDay('${key}')">` +
            `<span class="sch-num">${d}</span><span class="sch-sub">${notes.join('<br>')}</span></div>`;
    }
    document.getElementById('schedCalendar').innerHTML = html;
    renderScheduleDay(idx);
}

function renderScheduleDay(idx) {
    const panel = document.getElementById('schedDayPanel');
    if (!schedSelected) {
        panel.innerHTML = '<p class="sch-hint">Выберите день в календаре, чтобы закрыть или открыть часы приёма.</p>';
        return;
    }
    const slots = schedData.slots;
    const e = schedEntry(idx || schedIndex(), schedSelected);
    const info = schedDayInfo(e, slots);
    const past = schedSelected < schedTodayKey();
    const [y, m, d] = schedSelected.split('-').map(Number);
    const title = `${d} ${SCHED_MONTHS_GEN[m - 1]}, ${SCHED_WEEKDAYS[new Date(y, m - 1, d).getDay()]}`;

    const cells = slots.map(t => {
        const booked = e.booked[t];
        const block = e.fullBlock || e.slotBlocks[t];
        const cls = ['sch-slot'];
        const notes = [];
        if (booked) {
            cls.push('booked');
            notes.push(escapeHtml(booked.client_name) + (booked.status === 'pending' ? ' (ждёт оплаты)' : ''));
        }
        if (block) {
            cls.push('blocked');
            notes.push('закрыто' + (block.reason ? ' · ' + escapeHtml(block.reason) : ''));
        }
        if (!booked && !block) { cls.push('free'); notes.push('свободно'); }
        return `<div class="${cls.join(' ')}" onclick="schedToggleSlot('${schedSelected}', '${t}')">` +
            `<span class="sch-slot-time">${t}</span><span class="sch-slot-note">${notes.join('<br>')}</span></div>`;
    }).join('');

    const hasBlocks = info.closed || info.blockedCount > 0;
    const controls = past
        ? '<p class="sch-hint">Это прошедший день, расписание менять нельзя.</p>'
        : `<p class="sch-hint">Нажмите на час, чтобы закрыть или открыть его. Занятые записью часы закрыть нельзя.</p>
           <div class="sch-form-row">
               <input type="text" id="schedReason" class="sch-input sch-wide" maxlength="200" placeholder="Причина закрытия (необязательно)">
               ${info.closed ? '' : '<button class="btn btn-small" onclick="schedCloseDay()">Закрыть весь день</button>'}
               ${hasBlocks ? '<button class="btn btn-small btn-danger" onclick="schedOpenDay()">Открыть весь день</button>' : ''}
           </div>`;

    panel.innerHTML = `<h3 class="sch-h3">${title}</h3><div class="sch-slots">${cells}</div>${controls}`;
}

function schedSelectDay(key) {
    schedSelected = key;
    renderSchedule();
}

async function schedChangeMonth(delta) {
    const next = new Date(schedYear, schedMonth + delta, 1);
    schedYear = next.getFullYear();
    schedMonth = next.getMonth();
    schedSelected = null;
    await loadSchedule();
}

function schedReasonValue() {
    const input = document.getElementById('schedReason');
    return input ? input.value.trim() : '';
}

// Создание закрытий. Если на это время уже есть записи, бэкенд вернёт список, и мы спросим, закрывать ли всё равно.
async function schedPostBlocks(payload) {
    try {
        return await apiFetch('/admin/blocks', { method: 'POST', body: JSON.stringify(payload) });
    } catch (err) {
        if (err.status === 409 && err.data && err.data.need_force) {
            const lines = (err.data.conflicts || []).map(c => `• ${c.when}: ${c.client}`).join('\n');
            if (!confirm(`${err.data.error}\n\n${lines}\n\nЗакрытие не отменяет эти записи, клиенты придут по записи. Закрыть всё равно?`)) return null;
            return await apiFetch('/admin/blocks', {
                method: 'POST',
                body: JSON.stringify(Object.assign({}, payload, { force: true })),
            });
        }
        throw err;
    }
}

async function schedRun(action, errorText) {
    if (schedBusy) return;
    schedBusy = true;
    try {
        await action();
    } catch (err) {
        alert(errorText + ': ' + err.message);
    } finally {
        schedBusy = false;
        await loadSchedule();
    }
}

function schedToggleSlot(date, time) {
    if (schedBusy) return;
    if (date < schedTodayKey()) { alert('Это прошедший день, расписание менять нельзя.'); return; }

    const e = schedEntry(schedIndex(), date);
    if (e.fullBlock) {
        alert('Весь день закрыт. Чтобы открыть отдельные часы, сначала нажмите «Открыть весь день».');
        return;
    }
    const block = e.slotBlocks[time];
    if (!block && e.booked[time]) {
        alert('Этот час занят записью. Чтобы освободить его, отмените или перенесите запись на вкладке «Записи клиентов».');
        return;
    }
    return schedRun(async () => {
        if (block) {
            await apiFetch('/admin/blocks/' + block.id, { method: 'DELETE' });
        } else {
            await schedPostBlocks({ date, slot_times: [time], reason: schedReasonValue() });
        }
    }, 'Не удалось изменить расписание');
}

function schedCloseDay() {
    if (!schedSelected) return;
    const date = schedSelected;
    return schedRun(() => schedPostBlocks({ date, reason: schedReasonValue() }), 'Не удалось закрыть день');
}

function schedOpenDay() {
    if (!schedSelected) return;
    const date = schedSelected;
    return schedRun(
        () => apiFetch(`/admin/blocks?from=${date}&to=${date}`, { method: 'DELETE' }),
        'Не удалось открыть день'
    );
}

function schedReadPeriod() {
    const from = document.getElementById('schedFrom').value;
    const to = document.getElementById('schedTo').value || from;
    if (!from) { alert('Укажите дату «с».'); return null; }
    if (to < from) { alert('Дата «по» раньше даты «с».'); return null; }

    const weekdays = [...document.querySelectorAll('.sch-wd:checked')].map(el => Number(el.value));
    if (weekdays.length === 0) { alert('Выберите хотя бы один день недели.'); return null; }

    const payload = { from, to, reason: document.getElementById('schedPeriodReason').value.trim() };
    if (weekdays.length < 7) payload.weekdays = weekdays;

    const mode = document.querySelector('input[name="schedMode"]:checked').value;
    if (mode === 'hours') {
        const hourFrom = document.getElementById('schedHourFrom').value;
        const hourTo = document.getElementById('schedHourTo').value;
        const times = SCHED_SLOTS.filter(t => t >= hourFrom && t < hourTo);
        if (times.length === 0) { alert('Время «до» должно быть позже времени «с».'); return null; }
        if (times.length < SCHED_SLOTS.length) payload.slot_times = times;
    }
    return payload;
}

function schedClosePeriod() {
    const payload = schedReadPeriod();
    if (!payload) return;
    return schedRun(async () => {
        const res = await schedPostBlocks(payload);
        if (res) {
            alert(`Закрыто: ${res.created}` + (res.skipped ? `. Уже было закрыто: ${res.skipped}.` : '.'));
        }
    }, 'Не удалось закрыть период');
}

function schedOpenPeriod() {
    const from = document.getElementById('schedFrom').value;
    const to = document.getElementById('schedTo').value || from;
    if (!from) { alert('Укажите дату «с».'); return; }
    if (to < from) { alert('Дата «по» раньше даты «с».'); return; }
    if (!confirm(`Открыть всё, что закрыто с ${from} по ${to}? Записи клиентов это не затронет.`)) return;
    return schedRun(async () => {
        const res = await apiFetch(`/admin/blocks?from=${from}&to=${to}`, { method: 'DELETE' });
        alert(`Открыто закрытий: ${res.deleted}.`);
    }, 'Не удалось открыть период');
}
