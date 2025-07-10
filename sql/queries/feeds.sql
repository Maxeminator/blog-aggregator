-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: ListFeedsWithUsers :many
SELECT feeds.name, feeds.url, users.name FROM feeds
JOIN users ON feeds.user_id=users.id;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;