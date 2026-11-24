"""Calculate point-in-time market studies from verified daily API observations."""
import gzip
import json
import re
import subprocess
from collections import defaultdict, deque
from concurrent.futures import ThreadPoolExecutor
from decimal import Decimal

from build_market_review import SCOPE
from sync_market_review import SERVER

PLATES = "https://api.zizizaizai.com/v3/market/plates/14"
PATTERNS = ("上涨中继", "EMA5 小平台", "EMA5 大平台", "EMA13 回踩", "超跌回踩")


def record(work, importer, row, kind, data):
    path = work / f"{row['date']}-{kind}.json.gz"
    data["rule"] = {"patterns":"pullback-v1","style":"tdx-ema-csi20-v1","sectors":"ths-industry-high100-v1"}[kind]
    doc = dict(date=row["date"], batch=row["batch"], kind=kind, data=data)
    path.write_bytes(gzip.compress(json.dumps(doc, ensure_ascii=False, allow_nan=False, separators=(",", ":")).encode()))
    if importer is not None:
        subprocess.run([str(importer), "-study", str(path)], cwd=SERVER, check=True)


def candidate_evidence(symbol, name, bars):
    # A pattern ends at the selected close. Suspensions break a platform.
    today = bars[-1]
    i,day,o,h,l,c,amount,e5,e13=today
    if len(bars)<2 or h<=l or not e13:
        return []
    out=[]
    for kind in PATTERNS:
        options=[]
        for j,a in enumerate(bars[:-1]):
            step=i-a[0]
            if step!=len(bars)-1-j:
                continue
            after=bars[j+1:]
            if ((kind=="上涨中继" and step!=1) or (kind=="EMA5 小平台" and not 2<=step<=4)
                or (kind=="EMA5 大平台" and not 5<=step<=7)
                or (kind in ("EMA13 回踩","超跌回踩") and not 2<=step<=12)):
                continue
            _,anchor,ao,ah,al,ac,_,_,_=a
            body=100*(ac/ao-1)
            if body<1 or ah<=al:
                continue
            support = al if kind=="超跌回踩" else e13 if kind=="EMA13 回踩" else e5
            near=abs(100*(l/support-1))
            width=100*(max(b[3] for b in after)/min(b[4] for b in after)-1)
            if kind.startswith("EMA5") and (c<ao or c<e5):
                continue
            if kind=="EMA13 回踩" and (c<e13 or any(b[4]<=b[8] for b in after[:-1])):
                continue
            if kind=="超跌回踩" and (len(after)<3 or any(b[5]>=b[8] for b in after[-3:-1]) or c<=o):
                continue
            shrink=(h-l)/o/((ah-al)/ao)
            options.append(dict(kind=kind,symbol=symbol,name=name,anchor=anchor,days=step,big=round(body,4),
                                near=round(near,4),width=round(width,4),doji=round(100*abs(c-o)/(h-l),4),
                                shrink=round(shrink,5),close=round(c,4),ema5=round(e5,4),ema13=round(e13,4),
                                open=round(o,4),high=round(h,4),low=round(l,4),amount=round(amount,1),signalDate=day))
        if options:
            # The reference is the strongest bullish body in this pattern's structural window.
            out.append(max(options,key=lambda x:(x["big"],x["anchor"])))
    return out


def calculate_evidence(indices, targets, signal_days, traded, extra, names, partitions):
    peaks=defaultdict(lambda:deque(maxlen=100))
    closes=defaultdict(lambda:deque(maxlen=13))
    candles=defaultdict(lambda:deque(maxlen=13))
    averages={}
    prior_signals = {}
    for day,prices in partitions:
        eligible=set()
        hits=set()
        candidates=[]
        current_signals=defaultdict(list)
        di=indices[day]
        for s,(high,close,factor) in prices.items():
            if s not in traded or day<traded[s][0]:
                continue
            h=float(high*factor)
            c=float(close*factor)
            window=peaks[s]
            window.append(high*factor)
            if day in targets and len(window)==100:
                eligible.add(s)
                if window[-1]>max(list(window)[:-1]): hits.add(s)
            recent=closes[s]
            recent.append(c)
            for period in (5,13):
                key=(s,period)
                if key in averages: averages[key]+=2/(period+1)*(c-averages[key])
                elif len(recent)==period: averages[key]=sum(recent)/period
            if day not in extra or (s,13) not in averages: continue
            op,lo,amt=extra[day][s]
            if not (lo<=min(op,close)<=max(op,close)<=high):
                raise ValueError(f"invalid OHLC {s} {day}")
            bars=candles[s]
            bars.append((di,day,float(op*factor),h,float(lo*factor),c,float(amt)*1000,averages[s,5],averages[s,13]))
            if day in targets:
                for old,trigger_level in prior_signals.get(s,[]):
                    if high*factor>trigger_level:
                        candidates.append(dict(old,triggerDate=day))
            if day in signal_days:
                for r in candidate_evidence(s,names.get(s,s),list(bars)):
                    for k in ("close","open","high","low","ema5","ema13"):
                        r[k]=round(r[k]/float(factor),4)
                    current_signals[s].append((r,high*factor))
                    candidates.append(r)
        prior_signals=current_signals
        if day in targets:
            expected={s for s,seq in traded.items() if day in seq}
            if expected!=prices.keys(): raise ValueError(f"study roster differs {day}")
            # Dolt amounts are in thousands of yuan. Sum each traded A share once.
            market_amount = None
            if expected == extra.get(day, {}).keys():
                market_amount = float(sum(bar[2] for bar in extra[day].values()) * 1000)
            yield day, eligible, hits, candidates, market_amount


def industry_data(path):
    raw = subprocess.check_output(["curl", "-fsS", "--compressed", "--max-time", "15",
                                   "--retry", "1", PLATES + path])
    return json.loads(raw)["data"]


def sector_snapshot(day, calendar, eligible, hits):
    boards = industry_data(f"/rank?date1={day}&limit=500")
    if not boards or len(boards) >= 500 or len({b["plate_code"] for b in boards}) != len(boards):
        raise ValueError(f"{day} 行业目录不完整")
    end = calendar.index(day)
    base, sessions = calendar[end-20], calendar[end-19:end+1]

    def sector(board):
        code = board["plate_code"]
        if board["date1"] != day or not re.fullmatch(r"\d{6}", code):
            raise ValueError("行业日期或代码错误")
        members = industry_data(f"/{code}/stocks?date={day}")
        active = set()
        for m in members:
            if m["plate_code"] != code or not m["time_in"]:
                raise ValueError("行业成员记录无效")
            if m["time_in"] <= day and (not m["time_out"] or m["time_out"] > day):
                symbol = m["stock_code"] + (".SH" if m["stock_code"].startswith("6") else ".SZ")
                if SCOPE.fullmatch(symbol):
                    active.add(symbol)
        if not active:
            raise ValueError(f"{day} 行业 {code} 缺少成份")
        rates = {}
        for item in industry_data(f"/trend?plate_code={code}&day_start={base}&day_end={day}"):
            stamp = item["date1"]
            rate = Decimal(str(item["rate"]))
            if item["plate_code"] != code or item["plate_type"] != 14 or stamp in rates or not rate.is_finite() or rate <= -100:
                raise ValueError(f"{code} {stamp} 行业日涨跌幅异常")
            rates[stamp] = rate
        strength = None
        if all(stamp in rates for stamp in sessions):
            growth = Decimal(1)
            for stamp in sessions:
                growth *= 1 + rates[stamp] / 100
            strength = float(round((growth - 1) * 100, 4))
        return dict(code=code, name=board["plate_name"], members=len(active), eligible=len(active & eligible),
                    hits=sorted(active & hits), strength=strength, strengthMethod="dailyReturns",
                    change=board.get("rate"), amount=board.get("trade_money"))

    with ThreadPoolExecutor(max_workers=4) as pool:
        rows = sorted(pool.map(sector, boards), key=lambda r: (-len(r["hits"]), r["code"]))
    return dict(base=base, rows=rows, unclassified=sorted(hits.difference(*(r["hits"] for r in rows))))
