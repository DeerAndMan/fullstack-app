-- ============================================================
-- 历史数据批量转化脚本
-- 用途：将 summary 表的历史推送数据一次性转化为 trade_daily_summary 日终快照
-- 使用方法：
--   mysql -u用户名 -p密码 数据库名 < scripts/sync_history.sql
--   或在 MySQL 客户端、Navicat、phpMyAdmin 等工具中直接执行
-- ============================================================

-- 1. 确保目标表存在（trade_daily_summary 已通过 AutoMigrate 创建，此处仅防御）
CREATE TABLE IF NOT EXISTS `trade_daily_summary` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `trade_date` date NOT NULL COMMENT '交易日（yyyy-MM-dd）',
  `dryk` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '当日盈亏',
  `drhz` decimal(16,8) NOT NULL DEFAULT '0.00000000' COMMENT '当日收益率',
  `zzc` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '总资产',
  `zxsz` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '最新市值',
  `zjye` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '资金余额',
  `ljyk` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '累计盈亏',
  `kyzj` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '可用资金',
  `intraday_max_dryk` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '盘中最大当日盈亏',
  `intraday_min_dryk` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '盘中最小当日盈亏',
  `position_count` int NOT NULL DEFAULT '0' COMMENT '持仓品种数',
  `snapshot_count` int NOT NULL DEFAULT '0' COMMENT '当日推送次数',
  `last_summary_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '日终快照来源 summary.id',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_trade_date` (`trade_date`),
  KEY `idx_dryk` (`dryk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易日终快照表';

-- 2. 批量写入或更新（ON DUPLICATE KEY UPDATE）
-- 每个交易日取 MAX(id) 对应的推送作为日终快照，同时计算日内极值
--
-- 注意：summary 历史数据存在两种推送格式，字段语义不同，必须分别适配：
--   * 老格式（约 2025-03 之前）：dryk / zzc 为 NULL，drhz 里存的其实是「当日盈亏金额」，
--     总资产落在 zsz 字段；
--   * 新格式：dryk 为当日盈亏金额、zzc 为期末总资产，drhz 才是真正的当日收益率（小数）。
-- 若照搬 drhz，老格式会把 1086.60 这类金额当成 108660% 的日收益率，
-- 净值连乘后累计收益率会放大成天文数字。
-- 这里以「dryk 是否有值」判别格式，老格式把 drhz 归位到 dryk、zsz 归位到 zzc，
-- 并把 drhz 置 0，由后端用「当日盈亏 / 当日期初资产」重算收益率。
INSERT INTO `trade_daily_summary` (
  `trade_date`,
  `dryk`,
  `drhz`,
  `zzc`,
  `zxsz`,
  `zjye`,
  `ljyk`,
  `kyzj`,
  `intraday_max_dryk`,
  `intraday_min_dryk`,
  `position_count`,
  `snapshot_count`,
  `last_summary_id`,
  `created_at`,
  `updated_at`
)
SELECT
  DATE(s.`date`)                                           AS trade_date,
  -- 当日盈亏：新格式取 dryk；老格式 dryk 缺失，drhz 存的才是盈亏金额
  COALESCE(
    CAST(NULLIF(TRIM(s.dryk), '') AS DECIMAL(20,4)),
    CAST(NULLIF(TRIM(s.drhz), '') AS DECIMAL(20,4)),
    0
  )                                                        AS dryk,
  -- 当日收益率：仅新格式的 drhz 是比例，老格式置 0 由后端兜底重算
  CASE WHEN NULLIF(TRIM(s.dryk), '') IS NULL THEN 0
       ELSE COALESCE(CAST(NULLIF(TRIM(s.drhz), '') AS DECIMAL(16,8)), 0)
  END                                                      AS drhz,
  -- 期末总资产：新格式取 zzc；老格式无 zzc，用 zsz（总市值）兜底
  COALESCE(
    CAST(NULLIF(TRIM(s.zzc), '') AS DECIMAL(20,4)),
    CAST(NULLIF(TRIM(s.zsz), '') AS DECIMAL(20,4)),
    0
  )                                                        AS zzc,
  COALESCE(CAST(NULLIF(TRIM(s.zxsz), '') AS DECIMAL(20,4)), 0) AS zxsz,
  COALESCE(CAST(NULLIF(TRIM(s.zjye), '') AS DECIMAL(20,4)), 0) AS zjye,
  COALESCE(s.ljyk, 0)                                      AS ljyk,
  COALESCE(s.kyzj, 0)                                      AS kyzj,
  COALESCE(d.max_dryk, 0)                                  AS intraday_max_dryk,
  COALESCE(d.min_dryk, 0)                                  AS intraday_min_dryk,
  COALESCE(s.num, 0)                                       AS position_count,
  d.cnt                                                    AS snapshot_count,
  s.id                                                     AS last_summary_id,
  NOW()                                                    AS created_at,
  NOW()                                                    AS updated_at
FROM summary s
JOIN (
  -- 子查询：每个交易日取最后一条推送（MAX(id)），同时算出日内盈亏极值
  SELECT DATE(`date`) AS d,
         MAX(id)      AS mid,
         COUNT(*)     AS cnt,
         -- 日内极值同样要兼容两种格式，统一取「盈亏金额」所在的那一列
         MAX(COALESCE(CAST(NULLIF(TRIM(dryk), '') AS DECIMAL(20,4)),
                      CAST(NULLIF(TRIM(drhz), '') AS DECIMAL(20,4)))) AS max_dryk,
         MIN(COALESCE(CAST(NULLIF(TRIM(dryk), '') AS DECIMAL(20,4)),
                      CAST(NULLIF(TRIM(drhz), '') AS DECIMAL(20,4)))) AS min_dryk
  FROM summary
  -- WHERE `date` >= '2023-01-01' AND `date` <= '2023-12-31'  -- 可选：指定同步区间，注释掉则全量同步
  GROUP BY DATE(`date`)
) d ON s.id = d.mid
ORDER BY trade_date ASC
ON DUPLICATE KEY UPDATE
  dryk               = VALUES(dryk),
  drhz               = VALUES(drhz),
  zzc                = VALUES(zzc),
  zxsz               = VALUES(zxsz),
  zjye               = VALUES(zjye),
  ljyk               = VALUES(ljyk),
  kyzj               = VALUES(kyzj),
  intraday_max_dryk  = VALUES(intraday_max_dryk),
  intraday_min_dryk  = VALUES(intraday_min_dryk),
  position_count     = VALUES(position_count),
  snapshot_count     = VALUES(snapshot_count),
  last_summary_id    = VALUES(last_summary_id),
  updated_at         = NOW();

-- 3. 显示执行结果
SELECT
  '✅ 历史数据同步完成' AS status,
  COUNT(*) AS total_days,
  MIN(trade_date) AS earliest_date,
  MAX(trade_date) AS latest_date
FROM trade_daily_summary;
