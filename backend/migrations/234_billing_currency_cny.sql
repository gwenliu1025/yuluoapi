-- 雨落 API 的业务余额、额度、售价、扣费和统计统一使用人民币。
-- 字段名中的历史 usd 后缀作为兼容 API/Schema 保留，不再代表实际币种。
ALTER TABLE IF EXISTS batch_image_jobs
    ALTER COLUMN currency SET DEFAULT 'CNY';

UPDATE batch_image_jobs
SET currency = 'CNY'
WHERE currency = 'USD'
  AND estimated_cost = 0
  AND COALESCE(hold_amount, 0) = 0
  AND COALESCE(actual_cost, 0) = 0;
