-- 1. 扩展 withdrawals 表
ALTER TABLE withdrawals 
ADD COLUMN required_approvals INT DEFAULT 2 NOT NULL,
ADD COLUMN current_approvals INT DEFAULT 0 NOT NULL,
ALTER COLUMN status TYPE VARCHAR(32),
ALTER COLUMN status SET DEFAULT 'pending_review';

-- 2. 新建 review 表
CREATE TABLE withdrawal_reviews (
    id SERIAL PRIMARY KEY,
    withdrawal_id BIGINT NOT NULL REFERENCES withdrawals(id), -- 使用 BigInt FK
    admin_id INT NOT NULL,
    status VARCHAR(16) NOT NULL, -- 'approve', 'reject'
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(withdrawal_id, admin_id) -- 核心约束
);
