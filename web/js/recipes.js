function createRecipeCard(recipe) {
    const card = document.createElement('div');
    card.className = 'col-md-4 recipe-card';

    const img = document.createElement('img');
    img.src = recipe.photoBase64;
    img.className = 'card-img-top';
    img.alt = recipe.title;

    const cardBody = document.createElement('div');
    cardBody.className = 'card-body';

    const title = document.createElement('h5');
    title.className = 'card-title';
    title.textContent = recipe.title;

    const ingredients = document.createElement('p');
    ingredients.className = 'card-text';

    if (recipe.ingredients && Array.isArray(recipe.ingredients)) {
        ingredients.textContent = recipe.ingredients.map(ingredient => ingredient.name).join(', ');
    } else {
        ingredients.textContent = 'No ingredients';
    }

    cardBody.appendChild(title);
    cardBody.appendChild(ingredients);

    card.appendChild(img);
    card.appendChild(cardBody);

    return card;
}

function decodeBase64Image(dataString) {
    const matches = dataString.match(/^data:([A-Za-z-+\/]+);base64,(.+)$/);
    const response = {};

    if (matches.length !== 3) {
        return new Error('Invalid input string');
    }

    response.type = matches[1];
    response.data = Buffer.from(matches[2], 'base64');

    return response;
}

function fetchRecipes() {
    const xhr = new XMLHttpRequest();
    xhr.open('POST', 'http://localhost:8088/recipes/list', true);
    xhr.setRequestHeader('Content-Type', 'application/json');

    xhr.onreadystatechange = function () {
        if (xhr.readyState === XMLHttpRequest.DONE) {
            if (xhr.status === 200) {
                const data = JSON.parse(xhr.responseText);
                const recipes = data.recipes || [];
                updateRecipeCards(recipes);
            } else {
                console.error('Error fetching recipes:', xhr.statusText);
            }
        }
    };

    xhr.send();
}

function updateRecipeCards(recipes) {
    const recipeContainer = document.getElementById('recipeContainer');
    recipeContainer.innerHTML = ''; // Clear existing cards

    recipes.forEach(recipe => {
        const card = createRecipeCard(recipe);
        recipeContainer.appendChild(card);
    });
}

document.addEventListener('DOMContentLoaded', fetchRecipes);

document.addEventListener("DOMContentLoaded", function () {
    var navContent = document.getElementById("navContent");

    // Получение данных пользователя из localStorage
    var user = JSON.parse(localStorage.getItem('user'));

    // Если пользователь авторизован, отображаем его имя в виде ссылки на профиль
    if (user) {
        var userId = user.userID;
        var username = user.username;

        var profileLink = document.createElement("a");
        profileLink.href = "http://localhost:8088/users/" + userId;
        profileLink.textContent = username;

        var profileListItem = document.createElement("li");
        profileListItem.classList.add("nav-item");
        profileListItem.appendChild(profileLink);

        navContent.appendChild(profileListItem);
    } else {
        // Если пользователь не авторизован, добавляем кнопку входа
        var loginLink = document.createElement("a");
        loginLink.href = "http://localhost:8088/auth/login";
        loginLink.textContent = "Login";

        var loginListItem = document.createElement("li");
        loginListItem.classList.add("nav-item");
        loginListItem.appendChild(loginLink);

        navContent.appendChild(loginListItem);
    }
});
