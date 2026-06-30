    if (!localStorage.getItem('cookiesAccepted')) {
      document.getElementById('cookie-banner').style.display = 'block';
    }
    function acceptCookies() {
      localStorage.setItem('cookiesAccepted', '1');
      document.getElementById('cookie-banner').style.display = 'none';
    }

    const MONTHS = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь', 'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'];
    const DAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
    let cDate = new Date(), selDay = null, selectedTimeStr = "";

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

        if (td < today && td.toDateString() !== today.toDateString()) {
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
    }

    function chMon(dir) { cDate.setMonth(cDate.getMonth() + dir); renderCal(); }
    renderCal();

    function selT(el) {
      if (el.classList.contains('busy')) return;
      document.querySelectorAll('.tsl.sel').forEach(e => e.classList.remove('sel'));
      el.classList.add('sel');
      selectedTimeStr = el.textContent;
      document.getElementById('timeLabel').textContent = '— ' + selectedTimeStr;
    }

    function selFmt(f) {
      document.getElementById('fmtOn').classList.toggle('on', f === 'on');
      document.getElementById('fmtOff').classList.toggle('on', f === 'off');
    }

    function selPay(el) {
      document.querySelectorAll('.pay-opt').forEach(e => e.classList.remove('on'));
      el.classList.add('on');
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

    function confirmFinalB() {
      document.getElementById('modalStep1').style.display = 'none';
      document.getElementById('modalStep2').style.display = 'block';
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
      // Идентификатор вашей группы и количество постов
      const vkDomain = 'ekostapsy';
      const count = 10;

      const accessToken = 'c1d048f0c1d048f0c1d048f0c6c291ad28cc1d0c1d048f0abeaa66c531f7cea58d5c6f8';

      if (accessToken === ' ') {
        console.warn('Новостная лента: Пожалуйста, укажите настоящий сервисный ключ доступа VK API.');
        return;
      }

      // Подключаем JSONP запрос к API ВКонтакте
      const script = document.createElement('script');
      script.src = `https://api.vk.com/method/wall.get?domain=${vkDomain}&count=${count}&v=5.131&access_token=${accessToken}&callback=parseVkNews`;
      document.body.appendChild(script);
    });

    document.addEventListener('DOMContentLoaded', function () {
      const vkDomain = 'ekostapsy';
      const count = 15; // Запрашиваем 15 записей для запаса, чтобы после фильтрации точно набралось 4 штуки
      const accessToken = ' '; // Замените на ваш реальный ключ доступа

      // Если ключ не изменен, включаем демонстрационные новости
      if (accessToken === ' ' || !accessToken) {
        console.warn('Новостная лента: Используются демонстрационные данные. Вставьте настоящий ключ API VK.');
        loadDemoNews();
        return;
      }

      // Если ключ есть — делаем запрос к ВК
      const script = document.createElement('script');
      script.src = `https://api.vk.com/method/wall.get?domain=${vkDomain}&count=${count}&v=5.131&access_token=${accessToken}&callback=parseVkNews`;
      document.body.appendChild(script);
    });

    // Функция для обработки реальных данных от ВКонтакте
    function parseVkNews(data) {
      if (!data || !data.response || !data.response.items) {
        console.error('Не удалось загрузить новости VK:', data);
        loadDemoNews();
        return;
      }

      const container = document.querySelector('.news-grid');
      if (!container) return;
      container.innerHTML = '';

      // ФИЛЬТР: Оставляем только те записи, у которых есть И текст, И картинка
      const validItems = data.response.items.filter(item => {
        const hasText = item.text && item.text.trim().length > 0;

        let hasPhoto = false;
        if (item.attachments) {
          hasPhoto = item.attachments.some(att => att.type === 'photo');
        }

        return hasText && hasPhoto; // Пост подходит, только если выполняются оба условия
      });

      // Если после фильтрации ничего не нашлось, показываем демо-записи, чтобы секция не была пустой
      if (validItems.length === 0) {
        console.warn('Новостная лента: Среди последних постов не найдено записей с текстом и фото одновременно.');
        loadDemoNews();
        return;
      }

      // Берем первые 4 подходящие записи из отфильтрованных
      validItems.slice(0, 4).forEach(item => {
        const dateObj = new Date(item.date * 1000);
        const dateStr = dateObj.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' }).replace(' г.', '');
        const fullText = item.text || '';
        const lines = fullText.split('\n').map(l => l.trim()).filter(l => l.length > 0);

        let title = 'Новая публикация';
        let desc = 'Читайте подробнее в нашем сообществе ВКонтакте.';

        if (lines.length > 0) {
          title = lines[0].length > 50 ? lines[0].substring(0, 50) + '...' : lines[0];
          if (lines.length > 1) desc = lines.slice(1).join(' ');
        }
        if (desc.length > 120) desc = desc.substring(0, 120) + '...';

        // Извлекаем URL фотографии
        let photoUrl = '';
        if (item.attachments) {
          const firstPhoto = item.attachments.find(att => att.type === 'photo');
          if (firstPhoto && firstPhoto.photo && firstPhoto.photo.sizes) {
            const sizes = firstPhoto.photo.sizes;
            const optimal = sizes.find(s => s.type === 'q' || s.type === 'x') || sizes[sizes.length - 1];
            if (optimal) photoUrl = optimal.url;
          }
        }

        renderCard(dateStr, title, desc, photoUrl);
      });
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


  