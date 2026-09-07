const API_URL = "http://localhost:8080";
const USER_ID = 2;

const booksContainer = document.getElementById("books");
const readingListContainer = document.getElementById("readingList");

const searchInput = document.getElementById("search");
const authorInput = document.getElementById("author");
const searchButton = document.getElementById("searchButton");

async function loadBooks() {
    const params = new URLSearchParams();

    if (searchInput.value.trim()) {
        params.append("search", searchInput.value.trim());
    }

    if (authorInput.value.trim()) {
        params.append("author", authorInput.value.trim());
    }

    const response = await fetch(
        `${API_URL}/books?${params.toString()}`
    );

    const books = await response.json();

    renderBooks(books);
}

function renderBooks(books) {
    booksContainer.innerHTML = "";

    for (const book of books) {
        const card = document.createElement("div");

        card.className = "book-card";

        card.innerHTML = `
            <h3>${book.title}</h3>
            <p>${book.author.name}</p>
            <p>${book.description ?? ""}</p>
            <button data-id="${book.id}">
                Добавить в список
            </button>
        `;

        const button = card.querySelector("button");

        button.addEventListener("click", () => {
            addToReadingList(book.id);
        });

        booksContainer.appendChild(card);
    }
}

async function loadReadingList() {
    const response = await fetch(
        `${API_URL}/users/${USER_ID}/reading-list`
    );

    const books = await response.json();

    readingListContainer.innerHTML = "";

    for (const book of books) {
        const card = document.createElement("div");

        card.className = "book-card";

        card.innerHTML = `
            <h3>${book.title}</h3>
            <p>${book.author.name}</p>
            <button data-id="${book.id}">
                Удалить
            </button>
        `;

        const button = card.querySelector("button");

        button.addEventListener("click", () => {
            removeFromReadingList(book.id);
        });

        readingListContainer.appendChild(card);
    }
}

async function addToReadingList(bookId) {
    const response = await fetch(
        `${API_URL}/users/${USER_ID}/reading-list`,
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                book_id: bookId
            })
        }
    );

    if (!response.ok) {
        const error = await response.json();
        alert(error.error);
        return;
    }

    loadReadingList();
}

async function removeFromReadingList(bookId) {
    const response = await fetch(
        `${API_URL}/users/${USER_ID}/reading-list/${bookId}`,
        {
            method: "DELETE"
        }
    );

    if (!response.ok) {
        alert("Не удалось удалить книгу");
        return;
    }

    loadReadingList();
}

searchButton.addEventListener("click", loadBooks);

loadBooks();
loadReadingList();