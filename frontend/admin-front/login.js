document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const login = document.getElementById('login').value;
    const password = document.getElementById('password').value;
    const errorEl = document.getElementById('error');
    const submitBtn = e.target.querySelector('button[type="submit"]');

    errorEl.classList.add('hidden');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Входим...';

    try {
        const response = await fetch('/api/v1/admin/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: login, password }),
        });

        const result = await response.json();

        if (!response.ok) {
            throw new Error(result.error || 'Неверный логин или пароль');
        }

        sessionStorage.setItem('adminToken', result.token);
        sessionStorage.setItem('loginTime', Date.now());

        window.location.href = '/admin/admin/';
    } catch (err) {
        errorEl.textContent = 'Неверный логин или пароль';
        errorEl.classList.remove('hidden');
        document.getElementById('password').value = '';
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Войти';
    }
});