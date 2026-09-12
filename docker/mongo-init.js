db = db.getSiblingDB("book_catalog");

db.users.updateOne(
    { username: "admin" },
    {
        $set: {
            id: 1,
            username: "admin",
            password_hash: "$2a$10$hr6osSvcRI2t.RrVlL8j2eDckZN6KpzSOjWE1pJ2X7h6r/cN0deuK",
            role: "admin"
        }
    },
    { upsert: true }
);

db.users.createIndex(
    { id: 1 },
    { unique: true }
);

db.users.createIndex(
    { username: 1 },
    { unique: true }
);

db.authors.createIndex(
    { id: 1 },
    { unique: true }
);

db.books.createIndex(
    { id: 1 },
    { unique: true }
);

db.books.createIndex(
    { author_id: 1 }
);

db.reading_list.createIndex(
    {
        user_id: 1,
        book_id: 1
    },
    { unique: true }
);