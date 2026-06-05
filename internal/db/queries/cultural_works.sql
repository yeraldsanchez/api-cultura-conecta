-- name: CreateCulturalWork :one
WITH inserted AS (
    INSERT INTO cultural_works (title, category_id)
    VALUES ($1, $2)
    RETURNING *
)
SELECT i.*, c.name as category_name FROM inserted i
JOIN categories c ON i.category_id = c.id
WHERE c.id = $2;

-- name: GetCulturalWorks :many
SELECT cw.*, c.name as category_name
FROM cultural_works cw
JOIN categories c ON cw.category_id = c.id;

