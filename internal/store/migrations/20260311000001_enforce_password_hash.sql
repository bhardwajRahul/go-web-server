-- +goose Up
-- Disable legacy passwordless accounts before enforcing the new invariant.
UPDATE users
SET password_hash = 'account-disabled-no-password',
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE password_hash IS NULL;

ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;

-- +goose Down
-- Restore nullable password hashes for legacy compatibility.
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
