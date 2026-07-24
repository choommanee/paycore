-- internal/repository/queries/merchant_user.sql
-- Dashboard human identities. Login resolves a user by (provider, subject); a
-- first-time identity is created together with its merchant by the service layer.

-- name: GetMerchantUserByOAuth :one
SELECT * FROM merchant_users
WHERE oauth_provider = $1 AND oauth_subject = $2;

-- name: GetMerchantUserByID :one
SELECT * FROM merchant_users WHERE id = $1;

-- name: CreateMerchantUser :one
INSERT INTO merchant_users (
    id, merchant_id, email, name, avatar_url, oauth_provider, oauth_subject
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: TouchMerchantUserLogin :exec
UPDATE merchant_users SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;
