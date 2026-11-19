# 历史数据同步脚本使用说明

## 脚本功能

`sync_history.sql` 将 `summary` 表的历史推送数据批量转化为 `trade_daily_summary` 日终快照表，供统计功能使用。

- 每个交易日取最后一条推送（`MAX(id)`）作为日终快照
- 自动计算日内盈亏极值
- 已存在的日期会被覆盖更新（`ON DUPLICATE KEY UPDATE`）
- **执行一次即可，后续新数据由后端 API 自动同步**

## ⚠️ 两种历史推送格式

`summary` 的历史数据存在两种格式，字段语义不同，脚本已分别适配：

| 格式 | 时间范围 | dryk | drhz | 总资产字段 |
| ---- | -------- | ---- | ---- | ---------- |
| 老格式 | 约 2025-03 之前 | NULL | **当日盈亏金额** | `zsz` |
| 新格式 | 约 2025-03 之后 | 当日盈亏金额 | 当日收益率（小数） | `zzc` |

如果照搬 `drhz`，老格式的 `1086.60` 会被当成 108660% 的日收益率，净值连乘后累计收益率、
最大回撤、年化波动率会全部放大成天文数字。脚本以「`dryk` 是否有值」判别格式：
老格式把 `drhz` 归位到 `dryk`、`zsz` 归位到 `zzc`，并把 `drhz` 置 0，
由后端用「当日盈亏 / 当日期初资产」重算收益率。

**曾用旧版脚本同步过的库需要重跑一次**（或在统计页点「同步快照」），否则历史快照里的
`dryk` / `drhz` / `zzc` 仍是错位的。

## 执行方式

### 方式一：命令行（推荐）

```bash
# 1. 进入 server 目录
cd server

# 2. 执行脚本（会提示输入密码）
mysql -h数据库地址 -u用户名 -p 数据库名 < scripts/sync_history.sql

# 本地示例（假设你的配置是 localhost:3306，数据库 energytest，用户 root）
mysql -h127.0.0.1 -uroot -p energytest < scripts/sync_history.sql
```

执行完成后会显示：
```
✅ 历史数据同步完成    total_days: 244    earliest_date: 2023-01-03    latest_date: 2024-12-20
```

### 方式二：图形化工具

**Navicat / phpMyAdmin / MySQL Workbench / 宝塔面板**

1. 打开 `scripts/sync_history.sql` 文件
2. 复制全部内容
3. 在工具的「查询」或「SQL」窗口粘贴
4. 点击「运行」或「执行」
5. 查看执行结果

### 方式三：MySQL 客户端内执行

```bash
# 1. 登录 MySQL
mysql -h数据库地址 -u用户名 -p 数据库名

# 2. 在 MySQL 提示符下执行
mysql> source /完整路径/server/scripts/sync_history.sql
```

## 指定同步区间（可选）

如果只想同步部分历史（比如只要 2023 年），编辑 `sync_history.sql` 里子查询中那行被注释掉的 `WHERE`，取消注释并修改日期：

```sql
-- 修改前（全量同步）
  -- WHERE `date` >= '2023-01-01' AND `date` <= '2023-12-31'

-- 修改后（只同步 2023 年）
  WHERE `date` >= '2023-01-01' AND `date` <= '2023-12-31'
```

## 验证结果

执行后可以查询验证：

```sql
-- 查看日终快照总数和日期范围
SELECT 
  COUNT(*) AS total_days,
  MIN(trade_date) AS earliest_date,
  MAX(trade_date) AS latest_date
FROM trade_daily_summary;

-- 查看最近 10 个交易日的快照
SELECT * FROM trade_daily_summary ORDER BY trade_date DESC LIMIT 10;
```

## 常见问题

### 1. 报错 `Table 'trade_daily_summary' doesn't exist`

**原因**：表还没创建（AutoMigrate 未执行）

**解决**：
- 启动一次后端服务（`make dev-server` 或 `make dev-server-remote`），AutoMigrate 会自动建表
- 或者直接执行 `sync_history.sql`，脚本开头的 `CREATE TABLE IF NOT EXISTS` 会创建表

### 2. 连接远程数据库

```bash
# 远程数据库示例
mysql -h8.138.xxx.xxx -P3306 -uenergytest -p energytest < scripts/sync_history.sql

# 密码含特殊字符时，用引号包裹
mysql -h8.138.xxx.xxx -uenergytest -p'your(pass!word)' energytest < scripts/sync_history.sql
```

### 3. 执行时间

- 1000 条 summary 记录 ≈ 1-2 秒
- 1 万条 ≈ 10 秒
- 10 万条 ≈ 1 分钟
- 具体取决于服务器性能和网络延迟

### 4. 重复执行

可以重复执行，已存在的交易日会被覆盖更新，不会重复插入。

## 后续维护

**脚本执行完后就不再需要了**，后续新推送的数据由后端 API 自动同步：

- 每次 `/api/v1/trade/index` 推送数据时，会自动刷新当天的日终快照
- 手动触发全量同步：前端统计页面点击「同步快照」按钮，或调用 `POST /api/v1/trade/stats/sync`
