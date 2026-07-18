if (!localStorage.getItem('cookiesAccepted')) {
  document.getElementById('cookie-banner').style.display = 'block';
}
function acceptCookies() {
  localStorage.setItem('cookiesAccepted', '1');
  document.getElementById('cookie-banner').style.display = 'none';
}

const MONTHS = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь', 'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'];
const DAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const SLOT_TIMES = ['10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00'];
let cDate = new Date(), selDay = null, selectedTimeStr = "";
let busyByDate = {};

function formatDateKey(y, m, d) {
  return `${y}-${String(m + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
}

async function fetchAvailability(y, m) {
  const monthParam = `${y}-${String(m + 1).padStart(2, '0')}`;
  try {
    const response = await fetch(`/api/v1/availability?month=${monthParam}`);
    busyByDate = await response.json();
  } catch (err) {
    console.error('Не удалось загрузить занятые слоты:', err);
    busyByDate = {};
  }
  renderCal();
}

function renderCal() {
  const y = cDate.getFullYear(), m = cDate.getMonth();
  document.getElementById('calM').textContent = MONTHS[m] + ' ' + y;
  const g = document.getElementById('calG'); g.innerHTML = '';

  DAYS.forEach(d => {
    const el = document.createElement('div');
    el.className = 'cald';
    el.textContent = d;
    g.appendChild(el);
  });

  const first = new Date(y, m, 1), startDay = (first.getDay() + 6) % 7, dim = new Date(y, m + 1, 0).getDate(), today = new Date();

  for (let i = 0; i < startDay; i++) {
    const el = document.createElement('div');
    el.className = 'calday emp';
    g.appendChild(el);
  }

  for (let d = 1; d <= dim; d++) {
    const el = document.createElement('div');
    el.className = 'calday';
    el.textContent = d;
    const td = new Date(y, m, d);
    const dateKey = formatDateKey(y, m, d);
    const isFull = (busyByDate[dateKey] || []).length >= SLOT_TIMES.length;

    if ((td < today && td.toDateString() !== today.toDateString()) || isFull) {
      el.classList.add('dis');
    } else if (td.toDateString() === today.toDateString()) {
      el.classList.add('tod');
      el.addEventListener('click', () => setDateClick(el, d, m, y));
    } else {
      el.addEventListener('click', () => setDateClick(el, d, m, y));
    }

    if (selDay && selDay.d === d && selDay.m === m && selDay.y === y) {
      el.classList.add('sel');
    }
    g.appendChild(el);
  }
}

function setDateClick(el, d, m, y) {
  document.querySelectorAll('.calday.sel').forEach(e => e.classList.remove('sel'));
  el.classList.add('sel');
  selDay = { d, m, y };
  document.getElementById('dateLabel').textContent = d + ' ' + MONTHS[m] + ' ' + y;
  renderTimeSlots(formatDateKey(y, m, d));
}

function renderTimeSlots(dateKey) {
  selectedTimeStr = '';
  document.getElementById('timeLabel').textContent = '';

  const busy = busyByDate[dateKey] || [];
  const grid = document.getElementById('timeGrid');
  grid.innerHTML = '';

  SLOT_TIMES.forEach(t => {
    const el = document.createElement('div');
    el.className = 'tsl';
    el.textContent = t;
    if (busy.includes(t)) {
      el.classList.add('busy');
    } else {
      el.addEventListener('click', () => selT(el));
    }
    grid.appendChild(el);
  });
}

function chMon(dir) {
  cDate.setMonth(cDate.getMonth() + dir);
  selDay = null;
  selectedTimeStr = '';
  document.getElementById('dateLabel').textContent = 'не выбрана';
  document.getElementById('timeLabel').textContent = '';
  document.getElementById('timeGrid').innerHTML = '<div style="grid-column: 1/-1; color: var(--text-muted); font-size: 13px;">Сначала выберите дату приёма</div>';
  fetchAvailability(cDate.getFullYear(), cDate.getMonth());
}

fetchAvailability(cDate.getFullYear(), cDate.getMonth());

function selT(el) {
  if (el.classList.contains('busy')) return;
  document.querySelectorAll('.tsl.sel').forEach(e => e.classList.remove('sel'));
  el.classList.add('sel');
  selectedTimeStr = el.textContent;
  document.getElementById('timeLabel').textContent = '— ' + selectedTimeStr;
}

let selectedFormat = 'online';
let selectedPayMode = 'full';

function selFmt(f) {
  selectedFormat = (f === 'on') ? 'online' : 'offline';
  document.getElementById('fmtOn').classList.toggle('on', f === 'on');
  document.getElementById('fmtOff').classList.toggle('on', f === 'off');
}

function selPay(el) {
  document.querySelectorAll('.pay-opt').forEach(e => e.classList.remove('on'));
  el.classList.add('on');
  selectedPayMode = el.textContent.includes('50%') ? 'prepay_50' : 'full';
}

function submitB() {
  const name = document.getElementById('bName').value.trim();
  const phone = document.getElementById('bPhone').value.trim();
  const email = document.getElementById('bEmail').value.trim();
  const sEl = document.getElementById('bService');
  const serviceText = sEl.options[sEl.selectedIndex].text;
  const serviceVal = sEl.value;
  const dateText = document.getElementById('dateLabel').textContent;
  const consentCheck = document.getElementById('consent-check');

  if (!name) { alert("Пожалуйста, заполните поле: Ваше имя."); document.getElementById('bName').focus(); return; }
  if (!phone || phone.length < 9) { alert("Пожалуйста, введите корректный номер телефона."); document.getElementById('bPhone').focus(); return; }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!email || !emailRegex.test(email)) { alert("Пожалуйста, введите корректный адрес электронной почты."); document.getElementById('bEmail').focus(); return; }

  if (!serviceVal) { alert("Пожалуйста, выберите нужную услугу из списка."); document.getElementById('bService').focus(); return; }
  if (!selDay) { alert("Пожалуйста, выберите дату на календаре формы."); return; }
  if (!selectedTimeStr) { alert("Пожалуйста, укажите подходящее время приема."); return; }
  if (!consentCheck || !consentCheck.checked) { alert('Пожалуйста, дайте согласие на обработку персональных данных и примите условия договора-оферты'); return; }

  document.getElementById('sumName').textContent = name;
  document.getElementById('sumPhone').textContent = phone;
  document.getElementById('sumEmail').textContent = email;
  document.getElementById('sumService').textContent = serviceText;
  document.getElementById('sumDateTime').textContent = dateText + " в " + selectedTimeStr;

  document.getElementById('modalStep1').style.display = 'block';
  document.getElementById('modalStep2').style.display = 'none';
  document.getElementById('mdl').classList.add('open');
}

async function confirmFinalB() {
  const name = document.getElementById('bName').value.trim();
  const phone = document.getElementById('bPhone').value.trim();
  const email = document.getElementById('bEmail').value.trim();
  const serviceId = document.getElementById('bService').value;

  const nameParts = name.split(' ').filter(Boolean);
  const firstName = nameParts[0] || name;
  const lastName = nameParts.slice(1).join(' ') || firstName;

  const [hh, mm] = selectedTimeStr.split(':').map(Number);
  const startsAtISO = new Date(selDay.y, selDay.m, selDay.d, hh, mm, 0).toISOString();

  const btn = document.getElementById('confirmBtn');
  btn.disabled = true;
  btn.textContent = 'Отправляем...';

  try {
    const response = await fetch('/api/v1/appointments', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        client: {
          first_name: firstName,
          last_name: lastName,
          patronym: '',
          phone: phone,
          email: email,
        },
        service_id: serviceId,
        format: selectedFormat,
        starts_at: startsAtISO,
        payment_mode: selectedPayMode,
      }),
    });

    const result = await response.json();

    if (!response.ok) {
      if (response.status === 409) {
        throw new Error('Это время уже занято, пожалуйста, выберите другое.');
      }
      throw new Error(result.error || 'Не удалось создать запись. Попробуйте ещё раз.');
    }

    if (result.payment_url) {
      window.location.href = result.payment_url;
      return;
    }

    document.getElementById('modalStep1').style.display = 'none';
    document.getElementById('modalStep2').style.display = 'block';
  } catch (err) {
    alert(err.message);
  } finally {
    btn.disabled = false;
    btn.textContent = 'Всё верно, подтвердить';
  }
}

function closeMdl() {
  document.getElementById('mdl').classList.remove('open');
}

document.getElementById('mdl').addEventListener('click', function (e) {
  if (e.target === this) closeMdl();
});

function openServiceModal(contentId, titleText) {
  // Получаем нужный скрытый текст и переносим его в модалку
  const fullContent = document.getElementById(contentId).innerHTML;
  document.getElementById('s-modal-body').innerHTML = fullContent;

  // Устанавливаем заголовок
  document.getElementById('s-modal-title').innerText = titleText;

  // Показываем окно
  document.getElementById('service-modal').classList.add('active');

  // Блокируем прокрутку основной страницы
  document.body.style.overflow = 'hidden';
}

function closeServiceModal(event) {
  // Закрываем, если клик был по крестику, по кнопке "Понятно" или по темному фону вокруг
  if (!event || event.target.classList.contains('s-modal-overlay') || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('service-modal').classList.remove('active');

    // Возвращаем прокрутку
    document.body.style.overflow = 'auto';
  }
}


document.addEventListener('DOMContentLoaded', function () {
  console.group('🚀 Загрузка новостей из ВКонтакте');
  
  const vkDomain = 'ekostapsy';
  const count = 15;
  
  console.log('📍 Домен сообщества:', vkDomain);
  console.log('🔢 Запрашиваем постов:', count);
  
  const script = document.createElement('script');
  const apiUrl = `https://api.vk.com/method/wall.get?domain=${vkDomain}&count=${count}&v=5.131&callback=parseVkNews`;
  
  console.log('🌐 URL запроса:', apiUrl);
  console.log('⏱️ Отправляю запрос...');
  
  script.src = apiUrl;
  
  // Обработка успешной загрузки скрипта
  script.onload = () => {
    console.log('✅ Скрипт JSONP успешно загружен в DOM');
  };
  
  // Обработка ошибки загрузки скрипта
  script.onerror = (error) => {
    console.error('❌ Критическая ошибка загрузки скрипта:', error);
    console.error('💡 Возможно, проблема с интернетом или API ВКонтакте недоступен');
    console.groupEnd();
    loadDemoNews();
  };
  
  document.body.appendChild(script);
  console.log('📎 Скрипт добавлен в document.body');
  console.groupEnd();
});

// Функция для обработки реальных данных от ВКонтакте
function parseVkNews(data) {
  console.group('📥 Обработка ответа от API ВКонтакте');
  
  console.log('📦 Сырые данные от API:', data);
  
  // Проверка на наличие ошибок API
  if (data && data.error) {
    console.error('❌ API ВКонтакте вернул ошибку:');
    console.error('  Код ошибки:', data.error.error_code);
    console.error('  Сообщение:', data.error.error_msg);
    console.error('  Тип:', data.error.error_msg);
    console.warn('⚠️ Загружаю демо-новости вместо реальных');
    console.groupEnd();
    loadDemoNews();
    return;
  }
  
  // Проверка на валидность структуры ответа
  if (!data || !data.response || !data.response.items) {
    console.error('❌ Некорректная структура ответа от API');
    console.error('  Есть data?', !!data);
    console.error('  Есть response?', !!(data && data.response));
    console.error('  Есть items?', !!(data && data.response && data.response.items));
    console.warn('⚠️ Загружаю демо-новости');
    console.groupEnd();
    loadDemoNews();
    return;
  }
  
  console.log('📊 Всего получено постов:', data.response.items.length);
  console.log('📋 Таблица постов:');
  console.table(data.response.items.map(item => ({
    id: item.id,
    date: new Date(item.date * 1000).toLocaleDateString('ru-RU'),
    text_length: item.text ? item.text.length : 0,
    has_attachments: !!item.attachments,
    attachments_count: item.attachments ? item.attachments.length : 0
  })));
  
  const container = document.querySelector('.news-grid');
  if (!container) {
    console.error('❌ Контейнер .news-grid не найден в DOM');
    console.groupEnd();
    return;
  }
  
  console.log('🧹 Очищаю контейнер новостей');
  container.innerHTML = '';
  
  console.group('🔍 Фильтрация постов (текст + фото)');
  
  const validItems = data.response.items.filter((item, index) => {
    const hasText = item.text && item.text.trim().length > 0;
    
    let hasPhoto = false;
    if (item.attachments) {
      hasPhoto = item.attachments.some(att => att.type === 'photo');
    }
    
    const isValid = hasText && hasPhoto;
    
    console.log(`  Пост #${item.id} (индекс ${index}): текст=${hasText}✓, фото=${hasPhoto}✓ → ${isValid ? '✅ подходит' : '❌ отсеян'}`);
    
    return isValid;
  });
  
  console.log(`📈 Результат фильтрации: ${validItems.length} из ${data.response.items.length} постов подходят`);
  console.groupEnd();
  
  if (validItems.length === 0) {
    console.warn('⚠️ Не найдено ни одного поста с текстом и фото одновременно');
    console.warn('💡 Возможно, в группе сейчас только текстовые или только фото-посты');
    console.groupEnd();
    loadDemoNews();
    return;
  }
  
  console.log(`🎯 Беру первые ${Math.min(4, validItems.length)} постов для отображения`);
  
  const postsToRender = validItems.slice(0, 4);
  
  console.log('📋 Посты для рендеринга:');
  console.table(postsToRender.map((item, i) => ({
    '№': i + 1,
    id: item.id,
    date: new Date(item.date * 1000).toLocaleDateString('ru-RU'),
    text_preview: item.text ? item.text.substring(0, 50) + '...' : ''
  })));
  
  console.group('🎨 Рендеринг карточек');
  
  postsToRender.forEach((item, index) => {
    console.log(`\n--- Карточка ${index + 1} ---`);
    
    const dateObj = new Date(item.date * 1000);
    const dateStr = dateObj.toLocaleDateString('ru-RU', { 
      day: 'numeric', 
      month: 'short', 
      year: 'numeric' 
    }).replace(' г.', '');
    
    console.log('📅 Дата:', dateStr);
    
    const fullText = item.text || '';
    const lines = fullText.split('\n').map(l => l.trim()).filter(l => l.length > 0);
    
    let title = 'Новая публикация';
    let desc = 'Читайте подробнее в нашем сообществе ВКонтакте.';
    
    if (lines.length > 0) {
      title = lines[0].length > 50 ? lines[0].substring(0, 50) + '...' : lines[0];
      console.log('📝 Заголовок:', title);
      
      if (lines.length > 1) {
        desc = lines.slice(1).join(' ');
        console.log(`📄 Строк в тексте: ${lines.length}`);
      }
    }
    
    if (desc.length > 120) {
      desc = desc.substring(0, 120) + '...';
    }
    
    console.log('💬 Описание:', desc.substring(0, 60) + '...');
    
    let photoUrl = '';
    if (item.attachments) {
      const firstPhoto = item.attachments.find(att => att.type === 'photo');
      if (firstPhoto && firstPhoto.photo && firstPhoto.photo.sizes) {
        const sizes = firstPhoto.photo.sizes;
        console.log('📸 Доступные размеры фото:', sizes.map(s => `${s.type}(${s.width}x${s.height})`).join(', '));
        
        const optimal = sizes.find(s => s.type === 'q' || s.type === 'x') || sizes[sizes.length - 1];
        if (optimal) {
          photoUrl = optimal.url;
          console.log('🖼️ Выбран размер:', optimal.type, `(${optimal.width}x${optimal.height})`);
        }
      }
    }
    
    if (!photoUrl) {
      console.warn('⚠️ Фото не найдено для этого поста');
    }
    
    renderCard(dateStr, title, desc, photoUrl);
    console.log('✅ Карточка успешно отрендерена');
  });
  
  console.groupEnd();
  console.log('🎉 Все новости успешно loaded!');
  console.groupEnd();
}

// Функция отрисовки одной карточки
function renderCard(date, title, desc, imgUrl) {
  const container = document.querySelector('.news-grid');
  if (!container) return;

  const card = document.createElement('a');
  card.href = 'https://vk.com/ekostapsy';
  card.target = '_blank';
  card.className = 'news-card';
  card.style.textDecoration = 'none';
  card.style.color = 'inherit';
  card.style.transition = 'transform 0.2s, border-color 0.2s';

  card.onmouseenter = () => card.style.borderColor = 'var(--primary)';
  card.onmouseleave = () => card.style.borderColor = 'var(--border-color)';

  card.innerHTML = `
        <div class="news-photo-placeholder" style="position: relative; overflow: hidden; background: #eee;">
          ${imgUrl
      ? `<img src="${imgUrl}" alt="Иллюстрация" style="position: absolute; top:0; left:0; width:100%; height:100%; object-fit: cover; object-position: center;">`
      : `<span style="font-size: 11px; color: var(--text-muted); display:flex; align-items:center; justify-content:center; height:100%;">[Фото публикации]</span>`
    }
        </div>
        <div class="news-body">
          <div class="news-date">${date}</div>
          <div class="news-title">${title}</div>
          <div class="news-desc">${desc}</div>
        </div>
      `;
  container.appendChild(card);
}

// Функция для генерации демонстрационных новостей
function loadDemoNews() {
  const container = document.querySelector('.news-grid');
  if (!container) return;
  container.innerHTML = '';

  const demoData = [
    { date: 'Сегодня', title: 'Профессиональный кризис как точка старта', desc: 'Как понять, что вы упёрлись в невидимый потолок, и почему страх ошибки блокирует любые действия лидера...', img: 'https://images.unsplash.com/photo-1507537297725-24a1c029d3ca?auto=format&fit=crop&w=600&q=80' },
    { date: '18 июн 2026', title: 'Синдром самозванца у топ-менеджеров', desc: 'Успешные кейсы, за которыми скрывается внутреннее сомнение. Разбираем инструменты бережной трансформации состояний...', img: 'https://images.unsplash.com/photo-1454165804606-c3d57bc86b40?auto=format&fit=crop&w=600&q=80' },
    { date: '10 июн 2026', title: 'Психотерапия или коучинг: что выбрать?', desc: 'В чём ключевое различие методов и почему их сочетание даёт мощный внутренний импульс для долгосрочных изменений...', img: 'https://images.unsplash.com/photo-1517245386807-bb43f82c33c4?auto=format&fit=crop&w=600&q=80' },
    { date: '01 июн 2026', title: 'Анонс программы трансформации «Новая глава»', desc: 'Старт индивидуальной стратегической работы по переработке сложного профессионального опыта в новые достижения...', img: 'https://images.unsplash.com/photo-1515378791036-0648a3ef77b2?auto=format&fit=crop&w=600&q=80' }
  ];

  demoData.forEach(item => renderCard(item.date, item.title, item.desc, item.img));
}
function openPolicyModal() {
  document.getElementById('policy-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closePolicyModal(event) {
  if (!event || event.target.id === 'policy-modal' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('policy-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}

function openPersDataModal() {
  document.getElementById('pers-data').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closePersDataModal(event) {
  if (!event || event.target.id === 'pers-data' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('pers-data').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}

function openOfertaModal() {
  document.getElementById('oferta-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeOfertaModal(event) {
  if (!event || event.target.id === 'oferta-modal' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('oferta-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}

function openBackRuleModal() {
  document.getElementById('backrule-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeBackRuleModal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('backrule-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}

// Доки модальные окна
function openDoc1Modal() {
  document.getElementById('doc1-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeDoc1Modal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('doc1-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}


function openDoc2Modal() {
  document.getElementById('doc2-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeDoc2Modal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('doc2-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}



function openDoc3Modal() {
  document.getElementById('doc3-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeDoc3Modal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('doc3-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}

function openDoc4Modal() {
  document.getElementById('doc4-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeDoc4Modal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('doc4-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}


function openDoc5Modal() {
  document.getElementById('doc5-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeDoc5Modal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('doc5-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}



function openDoc6Modal() {
  document.getElementById('doc6-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeDoc6Modal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('doc6-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}

function openDoc7Modal() {
  document.getElementById('doc7-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeDoc7Modal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('doc7-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}







function openDevModal() {
  document.getElementById('dev-modal').classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeDevModal(event) {
  if (!event || event.target.id === 'back' || event.target.classList.contains('s-modal-close') || event.target.classList.contains('s-modal-btn')) {
    document.getElementById('dev-modal').classList.remove('active');
    document.body.style.overflow = 'auto';
  }
}