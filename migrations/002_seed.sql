INSERT INTO users (
    username,
    password_hash,
    role
)
VALUES (
    'admin',
    '$2a$10$hr6osSvcRI2t.RrVlL8j2eDckZN6KpzSOjWE1pJ2X7h6r/cN0deuK',
    'admin'
)
ON CONFLICT (username) DO UPDATE
SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role;