"""Point-in-time sector, style and pullback evidence for the review archive."""
import csv
import gzip
import json
import re
import struct
import subprocess
import zipfile
from collections import defaultdict, deque
from concurrent.futures import ThreadPoolExecutor
from decimal import Decimal
from pathlib import Path

from build_market_review import SCOPE, SOURCE
from sync_market_review import DOLT, SERVER, fetch, price_partitions, volume_history

TDX = "https://www.tdx.com.cn/products/data/data/vipdoc/tdxzs_day.zip"
PLATES = "https://api.zizizaizai.com/v3/market/plates/14"
THS = "https://d.10jqka.com.cn/v4/line/bk_"
PATTERNS = ("上涨中继", "EMA5 小平台", "EMA5 大平台", "EMA13 回踩", "超跌回踩")


def cached(work, name, url, *args):
    path = work / name
    if not path.exists():
        data = fetch(url, "--compressed", *args)
        path.write_bytes(data)
    return path.read_bytes()


def record(work, importer, row, kind, data):
    path = work / f"{row['date']}-{kind}.json.gz"
    data["rule"] = {"patterns":"pullback-v1","style":"tdx-ema-csi20-v1","sectors":"ths-industry-high100-v1"}[kind]
    doc = dict(date=row["date"], batch=row["batch"], kind=kind, data=data)
    path.write_bytes(gzip.compress(json.dumps(doc, ensure_ascii=False, allow_nan=False, separators=(",", ":")).encode()))
    subprocess.run([str(importer), "-study", str(path)], cwd=SERVER, check=True)


def style_history(work):
    archive = work / "tdx-index.zip"
    if not archive.exists():
        subprocess.run(["curl", "-fsSL", "--retry", "2", "--max-time", "900", TDX, "-o", str(archive)], check=True)
    result = {}
    with zipfile.ZipFile(archive) as package:
        for code in ("880865", "880003"):
            files = [f for f in package.namelist() if f.lower().endswith(f"sh{code}.day")]
            if len(files) != 1:
                raise ValueError(f"missing TDX index {code}")
            history = {}
            closes = []
            ema = {}
            for d, _, _, _, last, _, _, _ in struct.iter_unpack("<5If2I", package.read(files[0])):
                compact = str(d)
                day = f"{compact[:4]}-{compact[4:6]}-{compact[6:]}"
                close = last / 100
                if close <= 0 or (history and day <= next(reversed(history))):
                    raise ValueError(f"invalid TDX series {code} {day}")
                previous = closes[-1] if closes else None
                closes.append(close)
                for n in (5, 13):
                    if len(closes) == n:
                        ema[n] = sum(closes) / n
                    elif len(closes) > n:
                        ema[n] += 2 / (n + 1) * (close - ema[n])
                history[day] = {"code": code, "close": close, "previous": previous, "ema5": ema.get(5),
                                "ema13": ema.get(13), "change": 100 * (close / previous - 1) if previous else None}
            result[code] = history
    return result


def extra_bars(work, days, commit):
    result = defaultdict(dict)
    wanted = set(days)
    def unpack(reply):
        for group in reply["rows"]:
            if group["d"] not in wanted:
                continue
            for symbol, opening, low, amount in json.loads(group["stocks"],parse_float=Decimal,parse_int=Decimal):
                if SCOPE.fullmatch(symbol):
                    if not opening >= low > 0 or amount < 0:
                        raise ValueError(f"invalid OHLC {group['d']} {symbol}")
                    result[group["d"]][symbol] = (opening,low,amount)
    for cache in work.glob(f"ohl-{commit}-*.gz"):
        unpack(json.loads(gzip.decompress(cache.read_bytes())))
    days = [d for d in days if d not in result]
    pending = deque(days[i:i + 5] for i in range(0,len(days),5))
    while pending:
        chunk = pending.popleft()
        sql = (f"select tradedate as d,left(symbol,3) as grp,json_arrayagg(json_array(symbol,open,low,amount)) as stocks "
               f"from ts_a_stock_eod_price as of '{commit}' where tradedate between '{chunk[0]}' and '{chunk[-1]}' "
               "group by tradedate,left(symbol,3)")
        cache = work / f"ohl-{commit}-{chunk[0]}-{chunk[-1]}.gz"
        if cache.exists():
            reply = json.loads(gzip.decompress(cache.read_bytes()))
        else:
            raw = fetch(DOLT, "-G", "--data-urlencode", "q=" + sql)
            reply = json.loads(raw)
            if "deadline" in reply.get("query_execution_message","").lower() and len(chunk)>1:
                i=len(chunk)//2
                pending.extendleft((chunk[i:],chunk[:i]))
                continue
            if reply.get("query_execution_status") != "Success" or len(reply["rows"]) >= 1000:
                raise ValueError(f"invalid OHLC partition {chunk[0]}")
            cache.write_bytes(gzip.compress(raw))
        unpack(reply)
    return result


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


def evidence(work, qlib, calendar, targets, manifest):
    traded = volume_history(qlib,calendar,targets)
    first = min(seq[0] for seq in traded.values())
    days = calendar[calendar.index(first):calendar.index(targets[-1])+1]
    extra_start=max(0,calendar.index(targets[0])-13)
    signal_days = set(targets) | {calendar[calendar.index(d)-1] for d in targets}
    extra=extra_bars(work,calendar[extra_start:calendar.index(targets[-1])+1],manifest["dolt_commit"])
    master=work/"input"/"stocks_list.csv"
    if not master.exists():
        master.parent.mkdir(exist_ok=True)
        fetch(SOURCE+"/stocks/stocks_list.csv",output=master)
    with master.open(encoding="utf-8-sig") as f:
        names={r["symbol"]:r["name"] for r in csv.DictReader(f)}
    peaks=defaultdict(lambda:deque(maxlen=100))
    closes=defaultdict(lambda:deque(maxlen=13))
    candles=defaultdict(lambda:deque(maxlen=13))
    averages={}
    prior_signals = {}
    for day,prices in price_partitions(work,days,manifest["dolt_commit"]):
        eligible=set()
        hits=set()
        candidates=[]
        current_signals=defaultdict(list)
        di=calendar.index(day)
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
            yield day, eligible, hits, candidates


def sector_snapshot(work,day,calendar,eligible,hits):
    compact=day.replace("-","")
    rows=json.loads(cached(work,"boards-"+day,PLATES+f"/rank?date1={day}&limit=500"))["data"]
    if not rows or len(rows)>=500 or len({r["plate_code"] for r in rows})!=len(rows):
        raise ValueError(f"incomplete industry catalogue {day}")
    base=calendar[calendar.index(day)-20]
    def sector(board):
        code=board["plate_code"]
        if board["date1"]!=day or not re.fullmatch(r"\d{6}",code): raise ValueError("invalid sector date/code")
        src=f"{PLATES}/{code}/stocks?date={day}"
        members=json.loads(cached(work,f"members-{code}-{day}",src))["data"]
        active=set()
        for m in members:
            if m["plate_code"]!=code or not m["time_in"]: raise ValueError("invalid membership")
            if m["time_in"]<=day and (not m["time_out"] or m["time_out"]>day):
                s=m["stock_code"]+(".SH" if m["stock_code"].startswith("6") else ".SZ")
                if SCOPE.fullmatch(s): active.add(s)
        if not active: raise ValueError(f"empty industry {code} {day}")
        closes={}
        for year in {day[:4],base[:4]}:
            file=work/f"close-{code}-{year}"
            if not file.exists():
                url=f"{THS}{code}/01/{year}.js"
                proc=subprocess.run(["curl","-sSL","--retry","2","--max-time","45","-e","https://q.10jqka.com.cn/",url],capture_output=True,check=True)
                data=proc.stdout
                if data.startswith(f"quotebridge_v4_line_bk_{code}_01_{year}(".encode()):file.write_bytes(data)
            if file.exists():
                body=file.read_bytes().split(b"(",1)[1].rsplit(b")",1)[0]
                for line in json.loads(body.decode("gb18030"))["data"].split(";"):
                    fields=line.split(",")
                    if len(fields)>4: closes[fields[0]]=float(fields[4])
        before=closes.get(base.replace("-","")); now=closes.get(compact)
        return dict(code=code,name=board["plate_name"],members=len(active),eligible=len(active&eligible),
                    hits=sorted(active&hits),strength=round(100*(now/before-1),4) if before and now else None,
                    start=before,close=now)
    with ThreadPoolExecutor(max_workers=2) as pool: result=list(pool.map(sector,rows))
    result.sort(key=lambda x:(-len(x["hits"]),x["code"]))
    return dict(base=base,rows=result,unclassified=sorted(hits-set().union(*(set(r["hits"]) for r in result))))


def complete_sector_strength(work, snapshots, calendar):
    """Fill missing index closes using the same industry's published daily returns."""
    gaps = defaultdict(list)
    for day, snapshot in snapshots.items():
        for row in snapshot["rows"]:
            if row["strength"] is None:
                gaps[row["code"]].append((day, row))
    for code, entries in gaps.items():
        first = min(snapshots[day]["base"] for day, _ in entries)
        last = max(day for day, _ in entries)
        url = f"{PLATES}/trend?plate_code={code}&day_start={first}&day_end={last}"
        reply = json.loads(cached(work, f"rates-{code}-{first}-{last}", url), parse_float=Decimal)
        rates = {}
        for item in reply["data"]:
            day = item["date1"]
            if item["plate_code"] != code or item["plate_type"] != 14 or day in rates:
                raise ValueError(f"invalid industry return {code} {day}")
            rate = Decimal(item["rate"])
            if not rate.is_finite() or rate <= -100:
                raise ValueError(f"invalid daily rate {code} {day}")
            rates[day] = rate
        for day, row in entries:
            end = calendar.index(day)
            growth = Decimal(1)
            for session in calendar[end-19:end+1]:
                growth *= 1 + rates[session] / 100
            row["strength"] = float(round((growth - 1) * 100, 4))
            row["strengthMethod"] = "dailyReturns"


def index_closes(work, calendar, targets, commit):
    start=calendar[max(0,calendar.index(targets[0])-20)]
    sql=(f"select tradedate,close from ts_a_stock_eod_price as of '{commit}' where symbol='000300.SH' "
         f"and tradedate between '{start}' and '{targets[-1]}' order by tradedate")
    reply=json.loads(cached(work,f"csi-{commit}-{start}-{targets[-1]}",DOLT,"-G","--data-urlencode","q="+sql))
    if reply["query_execution_status"]!="Success" or len(reply["rows"])>=1000:
        raise ValueError("incomplete CSI 300 series")
    return {r["tradedate"]:float(r["close"]) for r in reply["rows"]}


def publish(work, importer, qlib, calendar, manifest, coverage):
    byday={r["date"]:r for r in coverage}
    targets=sorted(byday)
    styles=None
    prices={}
    for d in targets:
        if "style" in byday[d]["kinds"]:continue
        if styles is None:
            styles=style_history(work)
            prices=index_closes(work,calendar,targets,manifest["dolt_commit"])
        rows=[styles[c][d] for c in ("880865","880003")]
        if any(row["ema13"] is None for row in rows):raise ValueError(f"style data unavailable {d}")
        i=calendar.index(d)
        before=[prices[k] for k in calendar[i-20:i]]
        row=byday[d]
        old=row.get("prevEligible")
        divergence=None if not old or not row["eligible"] else (
            prices[d]>max(before) and row["above"]*old<row["prevAbove"]*row["eligible"])
        index=dict(close=prices[d],previous=before[-1],previous20High=max(before),divergence=divergence)
        record(work,importer,row,"style",dict(rows=rows,spread=rows[0]["change"]-rows[1]["change"],index=index))
    need=[d for d in targets if any(k not in byday[d]["kinds"] for k in ("patterns","sectors"))]
    if not need:return
    pending=[]
    for d,e,h,patterns in evidence(work,qlib,calendar,need,manifest):
        if "patterns" not in byday[d]["kinds"]:
            record(work,importer,byday[d],"patterns",dict(rows=patterns))
        if "sectors" not in byday[d]["kinds"]:
            pending.append((d,e,h))
    # Publish the latest sector ranking first, then fill older sessions.
    for d,e,h in reversed(pending):
        snapshot = sector_snapshot(work,d,calendar,e,h)
        complete_sector_strength(work, {d:snapshot}, calendar)
        record(work,importer,byday[d],"sectors",snapshot)
        print(d,"行业归档已发布",flush=True)
