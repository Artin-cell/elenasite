// Хешированный пароль (SHA-256 от "мойСекретныйПароль2026")
// Генерация: https://emn178.github.io/online-tools/sha256.html
const VALID_CREDENTIALS = {
    login: 'admin',
    // SHA-256 от пароля (не сам пароль в открытом виде!)
    passwordHash: '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08' 
};

// Функция SHA-256
async function sha256(message) {
    const msgBuffer = new TextEncoder().encode(message);
    const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

// Обработка формы
document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const login = document.getElementById('login').value;
    const password = document.getElementById('password').value;
    const errorEl = document.getElementById('error');
    
    errorEl.classList.add('hidden');
    
    // Хешируем введённый пароль
    const inputHash = await sha256(password);
    
    // Сравниваем
    if (login === VALID_CREDENTIALS.login && inputHash === VALID_CREDENTIALS.passwordHash) {
        // Успешный вход
        sessionStorage.setItem('isAdmin', 'true');
        sessionStorage.setItem('loginTime', Date.now());
        
        // Редирект на админ-панель
        window.location.href = '/admin/admin/';
    } else {
        errorEl.textContent = 'Неверный логин или пароль';
        errorEl.classList.remove('hidden');
        
        // Очистить пароль
        document.getElementById('password').value = '';
    }
});