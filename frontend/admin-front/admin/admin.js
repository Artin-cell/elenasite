// Переключение вкладок
        function switchTab(event, tabId) {
            // Убираем активный класс у всех вкладок
            document.querySelectorAll('.admin-tab').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // Скрываем все контенты
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.remove('active');
            });
            
            // Добавляем активный класс к выбранной вкладке
            event.target.classList.add('active');
            
            // Показываем соответствующий контент
            document.getElementById(tabId).classList.add('active');
        }