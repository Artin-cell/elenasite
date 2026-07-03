// Хешированный пароль

const VALID_CREDENTIALS = {
    login: 'admin',
    // SHA-256 от пароля (не сам пароль в открытом виде!)
    passwordHash: 'beb55a2c10325def8568c9797526881558b8915ab7191710a23422f6349dcbf7' 
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