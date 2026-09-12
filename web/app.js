const DATABASES = {
    postgres: "http://localhost:8080",
    mongo: "http://localhost:8081"
};

let currentAPI = localStorage.getItem("api") || "rest";
let currentDB = localStorage.getItem("db") || "postgres";
let currentUser = null;
let authorsCache = [];
let readingListIds = new Set();

const apiSelect = document.getElementById("apiSelect");
const dbSelect = document.getElementById("dbSelect");
const statusElement = document.getElementById("status");
const authSection = document.getElementById("authSection");
const userSection = document.getElementById("userSection");
const usernameInput = document.getElementById("username");
const passwordInput = document.getElementById("password");
const loginButton = document.getElementById("loginButton");
const registerButton = document.getElementById("registerButton");
const logoutButton = document.getElementById("logoutButton");
const authMessage = document.getElementById("authMessage");
const currentUserElement = document.getElementById("currentUser");
const currentRoleElement = document.getElementById("currentRole");
const searchInput = document.getElementById("searchInput");
const authorInput = document.getElementById("authorInput");
const searchButton = document.getElementById("searchButton");
const resetButton = document.getElementById("resetButton");
const booksElement = document.getElementById("books");
const readingListElement = document.getElementById("readingList");
const adminSection = document.getElementById("adminSection");
const bookForm = document.getElementById("bookForm");
const bookIdInput = document.getElementById("bookId");
const bookTitleInput = document.getElementById("bookTitle");
const bookAuthorSelect = document.getElementById("bookAuthor");
const bookDescriptionInput = document.getElementById("bookDescription");
const bookFormTitle = document.getElementById("bookFormTitle");
const cancelBookEdit = document.getElementById("cancelBookEdit");
const authorForm = document.getElementById("authorForm");
const authorIdInput = document.getElementById("authorId");
const authorNameInput = document.getElementById("authorName");
const authorFormTitle = document.getElementById("authorFormTitle");
const cancelAuthorEdit = document.getElementById("cancelAuthorEdit");
const authorsElement = document.getElementById("authors");
const messageElement = document.getElementById("message");

function baseURL() {
    return DATABASES[currentDB];
}

function tokenKey() {
    return `book_catalog_token_${currentDB}`;
}

function getToken() {
    return localStorage.getItem(tokenKey());
}

function authHeaders() {
    const token = getToken();
    return token ? { Authorization: `Bearer ${token}` } : {};
}

function isAdmin() {
    return currentUser && currentUser.role === "admin";
}

function setMessage(text, type = "") {
    messageElement.textContent = text;
    messageElement.className = type;
}

function setAuthMessage(text, type = "") {
    authMessage.textContent = text;
    authMessage.className = type;
}

async function readError(response, fallback) {
    try {
        const data = await response.json();
        return data.error || data.message || fallback;
    } catch {
        return fallback;
    }
}

async function graphql(query, variables = {}, withAuth = false) {
    const headers = { "Content-Type": "application/json" };

    if (withAuth) {
        Object.assign(headers, authHeaders());
    }

    const response = await fetch(`${baseURL()}/graphql`, {
        method: "POST",
        headers,
        body: JSON.stringify({ query, variables })
    });

    const result = await response.json();

    if (result.errors && result.errors.length > 0) {
        throw new Error(result.errors[0].message);
    }

    if (!response.ok) {
        throw new Error("Ошибка GraphQL запроса");
    }

    return result.data;
}

async function checkConnection() {
    statusElement.textContent = "Проверка соединения...";

    try {
        const response = await fetch(`${baseURL()}/health`);
        if (!response.ok) throw new Error();
        statusElement.textContent = "Сервер доступен";
    } catch {
        statusElement.textContent = "Сервер недоступен";
    }
}

function updateUserInterface() {
    if (currentUser) {
        authSection.classList.add("hidden");
        userSection.classList.remove("hidden");
        currentUserElement.textContent = currentUser.username;
        currentRoleElement.textContent = currentUser.role;
    } else {
        authSection.classList.remove("hidden");
        userSection.classList.add("hidden");
        currentUserElement.textContent = "";
        currentRoleElement.textContent = "";
    }

    if (isAdmin()) {
        adminSection.classList.remove("hidden");
        renderAuthors();
    } else {
        adminSection.classList.add("hidden");
    }
}

async function restoreSession() {
    currentUser = null;

    if (!getToken()) {
        updateUserInterface();
        return;
    }

    try {
        const response = await fetch(`${baseURL()}/auth/me`, {
            headers: authHeaders()
        });

        if (!response.ok) throw new Error();
        currentUser = await response.json();
    } catch {
        localStorage.removeItem(tokenKey());
    }

    updateUserInterface();
}

async function authenticate(mode) {
    const username = usernameInput.value.trim();
    const password = passwordInput.value;

    setAuthMessage("");

    if (!username || !password) {
        setAuthMessage("Введите логин и пароль.", "error");
        return;
    }

    try {
        const response = await fetch(`${baseURL()}/auth/${mode}`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || "Ошибка авторизации");
        }

        localStorage.setItem(tokenKey(), data.token);
        currentUser = data.user;
        passwordInput.value = "";

        updateUserInterface();
        await loadReadingList();
        await loadBooks();
    } catch (error) {
        setAuthMessage(error.message, "error");
    }
}

async function logout() {
    localStorage.removeItem(tokenKey());
    currentUser = null;
    readingListIds.clear();
    resetBookForm();
    resetAuthorForm();
    updateUserInterface();
    readingListElement.textContent = "Для работы со списком необходимо войти.";
    await loadBooks();
}

async function loadBooks() {
    booksElement.textContent = "Загрузка...";

    try {
        let books;

        if (currentAPI === "rest") {
            const params = new URLSearchParams();
            if (searchInput.value.trim()) params.set("search", searchInput.value.trim());
            if (authorInput.value.trim()) params.set("author", authorInput.value.trim());
            const suffix = params.toString() ? `?${params.toString()}` : "";

            const response = await fetch(`${baseURL()}/books${suffix}`);
            if (!response.ok) throw new Error("Не удалось получить книги");
            books = await response.json();
        } else {
            const data = await graphql(`
                query Books($search: String, $author: String) {
                    books(filter: { search: $search, author: $author }) {
                        id
                        title
                        description
                        author { id name }
                    }
                }
            `, {
                search: searchInput.value.trim() || null,
                author: authorInput.value.trim() || null
            });

            books = data.books;
        }

        renderBooks(books);
    } catch (error) {
        booksElement.textContent = error.message;
    }
}

function renderBooks(books) {
    if (!books || books.length === 0) {
        booksElement.textContent = "Книги не найдены.";
        return;
    }

    const table = document.createElement("table");
    table.innerHTML = `
        <thead>
            <tr>
                <th>ID</th>
                <th>Название</th>
                <th>Автор</th>
                <th>Описание</th>
                <th>Действия</th>
            </tr>
        </thead>
        <tbody></tbody>
    `;

    const tbody = table.querySelector("tbody");

    for (const book of books) {
        const row = document.createElement("tr");
        const idCell = document.createElement("td");
        const titleCell = document.createElement("td");
        const authorCell = document.createElement("td");
        const descriptionCell = document.createElement("td");
        const actionsCell = document.createElement("td");

        idCell.textContent = book.id;
        titleCell.textContent = book.title;
        authorCell.textContent = book.author.name;
        descriptionCell.textContent = book.description || "";
        actionsCell.className = "actions";

        if (currentUser) {
            if (readingListIds.has(String(book.id))) {
                actionsCell.append("В списке ");
            } else {
                const addButton = document.createElement("button");
                addButton.textContent = "В список";
                addButton.addEventListener("click", () => addToReadingList(book.id));
                actionsCell.appendChild(addButton);
            }
        }

        if (isAdmin()) {
            const editButton = document.createElement("button");
            editButton.textContent = "Изменить";
            editButton.addEventListener("click", () => editBook(book));

            const deleteButton = document.createElement("button");
            deleteButton.textContent = "Удалить";
            deleteButton.addEventListener("click", () => deleteBook(book.id));

            actionsCell.append(editButton, deleteButton);
        }

        row.append(idCell, titleCell, authorCell, descriptionCell, actionsCell);
        tbody.appendChild(row);
    }

    booksElement.innerHTML = "";
    booksElement.appendChild(table);
}

async function loadAuthors() {
    try {
        if (currentAPI === "rest") {
            const response = await fetch(`${baseURL()}/authors`);
            if (!response.ok) throw new Error();
            authorsCache = await response.json();
        } else {
            const data = await graphql(`query { authors { id name } }`);
            authorsCache = data.authors;
        }
    } catch {
        authorsCache = [];
    }

    fillAuthorSelect();
    if (isAdmin()) renderAuthors();
}

function fillAuthorSelect() {
    bookAuthorSelect.innerHTML = "";

    for (const author of authorsCache) {
        const option = document.createElement("option");
        option.value = author.id;
        option.textContent = author.name;
        bookAuthorSelect.appendChild(option);
    }
}

function renderAuthors() {
    if (!isAdmin()) return;

    if (authorsCache.length === 0) {
        authorsElement.textContent = "Авторов нет.";
        return;
    }

    const table = document.createElement("table");
    table.innerHTML = `
        <thead>
            <tr><th>ID</th><th>Имя</th><th>Действия</th></tr>
        </thead>
        <tbody></tbody>
    `;

    const tbody = table.querySelector("tbody");

    for (const author of authorsCache) {
        const row = document.createElement("tr");
        const idCell = document.createElement("td");
        const nameCell = document.createElement("td");
        const actionsCell = document.createElement("td");

        idCell.textContent = author.id;
        nameCell.textContent = author.name;
        actionsCell.className = "actions";

        const editButton = document.createElement("button");
        editButton.textContent = "Изменить";
        editButton.addEventListener("click", () => editAuthor(author));

        const deleteButton = document.createElement("button");
        deleteButton.textContent = "Удалить";
        deleteButton.addEventListener("click", () => deleteAuthor(author.id));

        actionsCell.append(editButton, deleteButton);
        row.append(idCell, nameCell, actionsCell);
        tbody.appendChild(row);
    }

    authorsElement.innerHTML = "";
    authorsElement.appendChild(table);
}

async function loadReadingList() {
    readingListIds.clear();

    if (!currentUser) {
        readingListElement.textContent = "Для работы со списком необходимо войти.";
        return;
    }

    try {
        let books;

        if (currentAPI === "rest") {
            const response = await fetch(`${baseURL()}/me/reading-list`, {
                headers: authHeaders()
            });

            if (!response.ok) throw new Error("Не удалось получить список чтения");
            books = await response.json();
        } else {
            const data = await graphql(`
                query {
                    readingList {
                        id
                        title
                        author { id name }
                    }
                }
            `, {}, true);

            books = data.readingList;
        }

        for (const book of books) readingListIds.add(String(book.id));
        renderReadingList(books);
    } catch (error) {
        readingListElement.textContent = error.message;
    }
}

function renderReadingList(books) {
    if (!books || books.length === 0) {
        readingListElement.textContent = "Список пуст.";
        return;
    }

    const table = document.createElement("table");
    table.innerHTML = `
        <thead>
            <tr><th>ID</th><th>Название</th><th>Автор</th><th>Действия</th></tr>
        </thead>
        <tbody></tbody>
    `;

    const tbody = table.querySelector("tbody");

    for (const book of books) {
        const row = document.createElement("tr");
        const idCell = document.createElement("td");
        const titleCell = document.createElement("td");
        const authorCell = document.createElement("td");
        const actionsCell = document.createElement("td");

        idCell.textContent = book.id;
        titleCell.textContent = book.title;
        authorCell.textContent = book.author.name;

        const removeButton = document.createElement("button");
        removeButton.textContent = "Удалить";
        removeButton.addEventListener("click", () => removeFromReadingList(book.id));
        actionsCell.appendChild(removeButton);

        row.append(idCell, titleCell, authorCell, actionsCell);
        tbody.appendChild(row);
    }

    readingListElement.innerHTML = "";
    readingListElement.appendChild(table);
}

async function addToReadingList(bookId) {
    try {
        if (currentAPI === "rest") {
            const response = await fetch(`${baseURL()}/me/reading-list`, {
                method: "POST",
                headers: { "Content-Type": "application/json", ...authHeaders() },
                body: JSON.stringify({ book_id: Number(bookId) })
            });

            if (!response.ok) {
                throw new Error(await readError(response, "Не удалось добавить книгу"));
            }
        } else {
            await graphql(`
                mutation AddBook($bookId: ID!) {
                    addBookToReadingList(bookId: $bookId)
                }
            `, { bookId: String(bookId) }, true);
        }

        await loadReadingList();
        await loadBooks();
    } catch (error) {
        setMessage(error.message, "error");
    }
}

async function removeFromReadingList(bookId) {
    try {
        if (currentAPI === "rest") {
            const response = await fetch(`${baseURL()}/me/reading-list/${bookId}`, {
                method: "DELETE",
                headers: authHeaders()
            });

            if (!response.ok) {
                throw new Error(await readError(response, "Не удалось удалить книгу"));
            }
        } else {
            await graphql(`
                mutation RemoveBook($bookId: ID!) {
                    removeBookFromReadingList(bookId: $bookId)
                }
            `, { bookId: String(bookId) }, true);
        }

        await loadReadingList();
        await loadBooks();
    } catch (error) {
        setMessage(error.message, "error");
    }
}

async function saveBook(event) {
    event.preventDefault();

    const id = bookIdInput.value;
    const authorId = bookAuthorSelect.value;
    const title = bookTitleInput.value.trim();
    const description = bookDescriptionInput.value.trim();

    if (!title || !authorId) return;

    try {
        if (currentAPI === "rest") {
            const url = id ? `${baseURL()}/books/${id}` : `${baseURL()}/books`;
            const response = await fetch(url, {
                method: id ? "PUT" : "POST",
                headers: { "Content-Type": "application/json", ...authHeaders() },
                body: JSON.stringify({
                    author_id: Number(authorId),
                    title,
                    description
                })
            });

            if (!response.ok) {
                throw new Error(await readError(response, "Не удалось сохранить книгу"));
            }
        } else {
            const input = { authorId: String(authorId), title, description };

            if (id) {
                await graphql(`
                    mutation UpdateBook($id: ID!, $input: UpdateBookInput!) {
                        updateBook(id: $id, input: $input) { id }
                    }
                `, { id: String(id), input }, true);
            } else {
                await graphql(`
                    mutation CreateBook($input: CreateBookInput!) {
                        createBook(input: $input) { id }
                    }
                `, { input }, true);
            }
        }

        resetBookForm();
        setMessage("Книга сохранена.", "success");
        await loadBooks();
    } catch (error) {
        setMessage(error.message, "error");
    }
}

function editBook(book) {
    bookIdInput.value = book.id;
    bookTitleInput.value = book.title;
    bookAuthorSelect.value = book.author.id;
    bookDescriptionInput.value = book.description || "";
    bookFormTitle.textContent = "Изменить книгу";
    cancelBookEdit.classList.remove("hidden");
}

function resetBookForm() {
    bookIdInput.value = "";
    bookTitleInput.value = "";
    bookDescriptionInput.value = "";
    if (bookAuthorSelect.options.length > 0) bookAuthorSelect.selectedIndex = 0;
    bookFormTitle.textContent = "Добавить книгу";
    cancelBookEdit.classList.add("hidden");
}

async function deleteBook(id) {
    if (!confirm("Удалить книгу?")) return;

    try {
        if (currentAPI === "rest") {
            const response = await fetch(`${baseURL()}/books/${id}`, {
                method: "DELETE",
                headers: authHeaders()
            });

            if (!response.ok) {
                throw new Error(await readError(response, "Не удалось удалить книгу"));
            }
        } else {
            await graphql(`
                mutation DeleteBook($id: ID!) {
                    deleteBook(id: $id)
                }
            `, { id: String(id) }, true);
        }

        setMessage("Книга удалена.", "success");
        await loadReadingList();
        await loadBooks();
    } catch (error) {
        setMessage(error.message, "error");
    }
}

async function saveAuthor(event) {
    event.preventDefault();

    const id = authorIdInput.value;
    const name = authorNameInput.value.trim();
    if (!name) return;

    try {
        if (currentAPI === "rest") {
            const url = id ? `${baseURL()}/authors/${id}` : `${baseURL()}/authors`;
            const response = await fetch(url, {
                method: id ? "PUT" : "POST",
                headers: { "Content-Type": "application/json", ...authHeaders() },
                body: JSON.stringify({ name })
            });

            if (!response.ok) {
                throw new Error(await readError(response, "Не удалось сохранить автора"));
            }
        } else {
            if (id) {
                await graphql(`
                    mutation UpdateAuthor($id: ID!, $input: UpdateAuthorInput!) {
                        updateAuthor(id: $id, input: $input) { id }
                    }
                `, { id: String(id), input: { name } }, true);
            } else {
                await graphql(`
                    mutation CreateAuthor($input: CreateAuthorInput!) {
                        createAuthor(input: $input) { id }
                    }
                `, { input: { name } }, true);
            }
        }

        resetAuthorForm();
        setMessage("Автор сохранен.", "success");
        await loadAuthors();
        await loadBooks();
    } catch (error) {
        setMessage(error.message, "error");
    }
}

function editAuthor(author) {
    authorIdInput.value = author.id;
    authorNameInput.value = author.name;
    authorFormTitle.textContent = "Изменить автора";
    cancelAuthorEdit.classList.remove("hidden");
}

function resetAuthorForm() {
    authorIdInput.value = "";
    authorNameInput.value = "";
    authorFormTitle.textContent = "Добавить автора";
    cancelAuthorEdit.classList.add("hidden");
}

async function deleteAuthor(id) {
    if (!confirm("Удалить автора?")) return;

    try {
        if (currentAPI === "rest") {
            const response = await fetch(`${baseURL()}/authors/${id}`, {
                method: "DELETE",
                headers: authHeaders()
            });

            if (!response.ok) {
                throw new Error(await readError(response, "Не удалось удалить автора"));
            }
        } else {
            await graphql(`
                mutation DeleteAuthor($id: ID!) {
                    deleteAuthor(id: $id)
                }
            `, { id: String(id) }, true);
        }

        setMessage("Автор удален.", "success");
        await loadAuthors();
        await loadBooks();
    } catch (error) {
        setMessage(error.message, "error");
    }
}

async function reload() {
    apiSelect.value = currentAPI;
    dbSelect.value = currentDB;
    setMessage("");

    await checkConnection();
    await restoreSession();
    await loadAuthors();
    await loadReadingList();
    await loadBooks();
}

apiSelect.addEventListener("change", async () => {
    currentAPI = apiSelect.value;
    localStorage.setItem("api", currentAPI);
    resetBookForm();
    resetAuthorForm();
    await loadAuthors();
    await loadReadingList();
    await loadBooks();
});

dbSelect.addEventListener("change", async () => {
    currentDB = dbSelect.value;
    localStorage.setItem("db", currentDB);
    currentUser = null;
    readingListIds.clear();
    resetBookForm();
    resetAuthorForm();
    await reload();
});

loginButton.addEventListener("click", () => authenticate("login"));
registerButton.addEventListener("click", () => authenticate("register"));
logoutButton.addEventListener("click", logout);
searchButton.addEventListener("click", loadBooks);
resetButton.addEventListener("click", async () => {
    searchInput.value = "";
    authorInput.value = "";
    await loadBooks();
});

searchInput.addEventListener("keydown", event => {
    if (event.key === "Enter") loadBooks();
});

authorInput.addEventListener("keydown", event => {
    if (event.key === "Enter") loadBooks();
});

bookForm.addEventListener("submit", saveBook);
cancelBookEdit.addEventListener("click", resetBookForm);
authorForm.addEventListener("submit", saveAuthor);
cancelAuthorEdit.addEventListener("click", resetAuthorForm);

reload();
