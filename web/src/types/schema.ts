import { z } from "zod";

// 基础响应类型
export const BaseResponseSchema = z.object({
  data: z.any(),
  code: z.number(),
  message: z.string(),
});

// 交易相关 Schema
export const TradeParamsSchema = z.object({
  startTime: z.string().min(1, "开始时间不能为空"),
  endTime: z.string().min(1, "结束时间不能为空"),
});

// 能源项目 Schema（对应单只证券的持仓/交易明细数据）
export const EnergyItemSchema = z.object({
  id: z.number(), // 主键ID，唯一标识单条证券明细记录
  datetime: z.string(), // 日期时间字符串，格式通常为"YYYY-MM-DD HH:MM:SS"，记录数据产生时间
  Zqmc: z.string(), // 证券名称，如股票名称（例：中孚实业）
  Cbjgex: z.string(), // 扩展成本价格，包含手续费、印花税等杂费后的综合成本价（字符串形式的数值）
  Cbjg: z.string(), // 成本价格，持仓的每股基础成本价（未包含杂费，字符串形式的数值）
  Ckcb: z.string(), // 持仓成本，持有该证券的总平均成本金额（字符串形式的数值）
  Ckcbj: z.string(), // 持仓成本价，持有该证券的每股平均成本价（字符串形式的数值）
  Ckyk: z.string(), // 持仓盈亏，当前持仓对应的浮动盈亏金额（字符串形式的数值）
  Ykbl: z.string(), // 盈亏比例，持仓盈亏占持仓成本的百分比（字符串形式的比例值，如"0.0167"代表1.67%）
  Dryk: z.string(), // 当日盈亏，当天因股价波动或交易产生的盈亏金额（字符串形式的数值）
  Drykbl: z.string(), // 当日盈亏比例，当日盈亏占相关成本的百分比（字符串形式的比例值）
  Cwbl: z.string(), // 持仓比例，该证券持仓市值占账户总市值的百分比（字符串形式的比例值）
  Djsl: z.string(), // 冻结数量，因委托未成交、配股等被临时冻结的股数（字符串形式的整数）
  Dqcb: z.string(), // 当前成本，当前持仓的最新计算成本（字符串形式的数值）
  Gddm: z.string(), // 股东代码，用户在证券市场的唯一股东身份标识（如A股股东代码）
  Gfmcdj: z.string(), // 股份买入登记数量，买入后已完成登记的股数（字符串形式的整数）
  Gfmrjd: z.string(), // 股份买入冻结数量，买入委托后等待成交的冻结股数（字符串形式的整数）
  Gfssmmce: z.string(), // 股份变动数量，记录股份增减（负数为减少，正数为增加，字符串形式的整数）
  Gfye: z.string(), // 股份余额，当前可用于交易的未冻结股数（字符串形式的整数）
  Jgbm: z.string(), // 机构编码，用户所属交易机构或营业部的编号
  Khdm: z.string(), // 客户代码，证券公司为用户分配的唯一客户编号
  Ksssl: z.string(), // 可售数量，当前可卖出的最大股数（字符串形式的整数）
  Kysl: z.string(), // 可用数量，当前可直接用于交易的股数（字符串形式的整数）
  Ljyk: z.string(), // 累计盈亏，从持有该证券开始到当前的总盈亏金额（字符串形式的数值）
  Market: z.string(), // 市场标识，证券所属交易市场（如"HA"代表沪A，"SA"代表深A）
  Mrssc: z.string(), // 买入委托数量，当前未成交的买入委托总股数（字符串形式的整数）
  Sssl: z.string(), // 卖出委托数量，当前未成交的卖出委托总股数（字符串形式的整数）
  Szjsbs: z.string(), // 数字结算标识，标记结算方式的编码（如"1"代表实时结算）
  Zjzh: z.string(), // 资金账号，用户在证券公司的资金账户编号，用于资金结算
  Zqdm: z.string(), // 证券代码，证券在交易所的唯一编码（如"600595"为中孚实业代码）
  Zqlx: z.string(), // 证券类型代码，数字标识证券类别（如"0"代表股票，"1"代表债券）
  Zqlxmc: z.string(), // 证券类型名称，对Zqlx的文字说明（如"股票"、"债券"）
  Zqsl: z.string(), // 证券数量，当前持有该证券的总股数（字符串形式的整数）
  Ztmc: z.string(), // 主力名称，该证券的主力资金或主力营业部名称
  Ztmr: z.string(), // 主力买入数据，主力资金对该证券的买入相关统计（字符串形式的数值）
  Zxjg: z.string(), // 最新价格，该证券当前的实时成交价格（字符串形式的数值）
  Zxsz: z.string(), // 最新市值，当前持仓股数×最新价格的总价值（字符串形式的数值）
  Bz: z.string(), // 币种标识，交易结算使用的货币（如"RMB"代表人民币）
});

// 交易项目 Schema（对应账户级别的交易汇总数据，包含多个证券的持仓明细）
export const TradeItemSchema = z.object({
  id: z.number(), // 主键ID，唯一标识单条交易汇总记录
  date: z.string(), // 日期字符串，格式通常为"YYYY-MM-DD"，记录交易所属日期
  drhz: z.string(), // 当日汇总，当天交易的汇总信息（可能为状态标识或摘要）
  dryk: z.string(), // 当日盈亏，账户当天整体的盈亏金额（字符串形式的数值）
  zxsz: z.string(), // 最新市值，账户内所有证券的当前总市值（字符串形式的数值）
  zzc: z.string(), // 总资产，账户内总资产（含证券市值+资金余额等，字符串形式的数值）
  RMBZzc: z.string(), // 人民币总资产，以人民币计价的账户总资产（字符串形式的数值）
  num: z.number(), // 数量，可能为持仓证券的总只数或交易笔数
  zsz: z.string(), // 总市值，账户内证券的总市值（与zxsz可能存在统计维度差异，字符串形式的数值）
  ccyk: z.string(), // 持仓盈亏，账户内所有持仓证券的总浮动盈亏（字符串形式的数值）
  stocks: z.string(), // 股票汇总，账户内股票类资产的汇总信息（可能为名称列表或统计值）
  zjye: z.string(), // 资金余额，账户内当前的可用资金余额（字符串形式的数值）
  positions: z.array(EnergyItemSchema), // 持仓列表，包含账户内所有证券的持仓明细（数组形式，元素为单只证券数据）
  djzj: z.number(), // 冻结资金，因未成交委托等被冻结的资金金额
  kqzj: z.number(), // 可用资金，账户内当前可用于交易的资金金额
  ljyk: z.number(), // 累计盈亏，账户从开户到当前的总盈亏金额
  kyzj: z.number(), // 可用资金（可能与kqzj一致，或为另一维度的可用资金统计）
  money_type: z.string(), // 货币类型，账户主要结算货币（如"CNY"代表人民币）
  totalsecMKval: z.string(), // 总证券市值，账户内所有证券的市值总和（字符串形式的数值，可能与zxsz/zzc存在计算逻辑差异）
});

// 交易列表响应 Schema - 直接返回数组
export const TradeListResponseSchema = z.array(TradeItemSchema);

// 交易摘要响应 Schema - 直接返回单个项目
export const TradeSummaryResponseSchema = TradeItemSchema;

// 包装的响应 Schema（如果需要的话）
export const WrappedTradeListResponseSchema = BaseResponseSchema.extend({
  data: z.array(TradeItemSchema),
});

export const WrappedTradeSummaryResponseSchema = BaseResponseSchema.extend({
  data: TradeItemSchema,
});

// 导出类型
export type TradeParams = z.infer<typeof TradeParamsSchema>;
export type EnergyItem = z.infer<typeof EnergyItemSchema>;
export type TradeItem = z.infer<typeof TradeItemSchema>;
export type TradeListResponse = z.infer<typeof TradeListResponseSchema>;
export type TradeSummaryResponse = z.infer<typeof TradeSummaryResponseSchema>;
export type BaseResponse = z.infer<typeof BaseResponseSchema>;
