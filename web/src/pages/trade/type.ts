export interface TradeItem {
  id: number;
  date: string;
  drhz: string;
  dryk: string;
  zxsz: string;
  zzc: string;
  RMBZzc: string; // 人民币总资产
  num: number;
  zsz: string;
  ccyk: string;
  stocks: string;
  zjye: string;
  positions: EnergyItem[];
  djzj: number;
  kqzj: number;
  ljyk: number;
  kyzj: number;
  money_type: string;
  totalsecMKval: string;
}

export interface EnergyItem {
  id: number; // 主键
  datetime: string;
  Zqmc: string;
  Cbjgex: string;
  Cbjg: string;
  Ckcb: string;
  Ckcbj: string;
  Ckyk: string;
  Ykbl: string;
  Dryk: string;
  Drykbl: string;
  Cwbl: string;
  Djsl: string;
  Dqcb: string;
  Gddm: string;
  Gfmcdj: string;
  Gfmrjd: string;
  Gfssmmce: string;
  Gfye: string;
  Jgbm: string;
  Khdm: string;
  Ksssl: string;
  Kysl: string;
  Ljyk: string;
  Market: string;
  Mrssc: string;
  Sssl: string;
  Szjsbs: string;
  Zjzh: string;
  Zqdm: string;
  Zqlx: string;
  Zqlxmc: string;
  Zqsl: string;
  Ztmc: string;
  Ztmr: string;
  Zxjg: string;
  Zxsz: string;
  Bz: string;
}
