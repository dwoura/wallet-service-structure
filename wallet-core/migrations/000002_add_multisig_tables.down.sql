DROP TABLE IF EXISTS withdrawal_reviews;

ALTER TABLE withdrawals
DROP COLUMN IF EXISTS required_approvals,
DROP COLUMN IF EXISTS current_approvals,
ALTER COLUMN status SET DEFAULT NULL; -- Revert default if possible, or leave it.
-- We usually don't shorten varchar length back to 20 to avoid truncation if data grew.
