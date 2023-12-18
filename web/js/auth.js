// script.js
document.addEventListener("DOMContentLoaded", function () {
    var user = JSON.parse(localStorage.getItem('user'));
    var tokenExpired = false;

    if (user && user.expiredAt) {
        var expiredAt = new Date(user.expiredAt);
        tokenExpired = expiredAt < new Date();
    }

    if (!user || tokenExpired) {
        // Перенаправляем на страницу входа, только если мы не уже находимся на странице входа
        if (!window.location.href.includes('login')) {
            window.location.href = 'login';
        }
    } else {
        console.log('Welcome, ' + user.username);
        // Ваш дополнительный код для работы с информацией о пользователе
    }
});

function sendData() {
    var email = document.getElementById("email").value;
    var password = document.getElementById("password").value;

    // Простая валидация email и password
    if (!validateEmail(email)) {
        showError("emailError", "Please enter a valid email address.");
        return;
    }

    if (!validatePassword(password)) {
        showError("passwordError", "Password must be at least 6 characters long.");
        return;
    }

    var data = {
        email: email,
        password: password
    };

    fetch('http://localhost:8088/auth/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Wrong credentials.');
            }
            return response.json();
        })
        .then(data => {
            console.log('Success:', data);

            // Преобразование строк в объекты Date
            var createdAt = new Date(data.createdAt);
            var expiredAt = new Date(data.expiredAt);

            // Сохранение токена и информации о пользователе в Local Storage
            localStorage.setItem('user', JSON.stringify({
                id: data.id,
                userID: data.userID,
                username: data.username,
                token: data.token,
                createdAt: createdAt,
                expiredAt: expiredAt,
            }));

            // Перенаправление на главную страницу
            window.location.href = 'http://localhost:8088/recipes/list';
        })
        .catch((error) => {
            console.error('Error:', error.message);
            showError("loginError", error.message);
        });
}

function validateEmail(email) {
    var emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

function validatePassword(password) {
    return password.length >= 6;
}

function showError(elementId, message) {
    document.getElementById(elementId).innerHTML = message;
}