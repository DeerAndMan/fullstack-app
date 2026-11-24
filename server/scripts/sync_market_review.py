"""Publish the selected A-share trading day from market data APIs."""
import argparse
import fcntl
import json
import os
import subprocess
import tempfile
from datetime import date, datetime
from pathlib import Path
from zoneinfo import ZoneInfo

from review_quote_sources import QuotesNotPublished, SOURCES, options

SERVER = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--refresh-date", type=date.fromisoformat)
    parser.add_argument("--provider", choices=SOURCES, default="public")
    parser.add_argument("--list-providers", action="store_true")
    parser.add_argument("--prewarm", action="store_true", help="cache historical windows for today's close")
    parser.add_argument("--core-only", action="store_true", help="publish only the selected API close")
    parser.add_argument("--scheduled", action="store_true", help="skip market holidays and completed closes")
    args = parser.parse_args()
    if args.list_providers:
        print(json.dumps(options()))
        return
    reason = next((o["reason"] for o in options() if o["id"] == args.provider), "需要安装 BaoStock SDK")
    if reason or not args.refresh_date:
        parser.error(reason or "API 更新需要指定 --refresh-date")
    if (args.prewarm or args.scheduled) and args.provider != "public":
        parser.error("自动收盘任务需要使用当天的公开行情接口")
    now = datetime.now(ZoneInfo("Asia/Shanghai"))
    end = args.refresh_date
    if end > now.date():
        parser.error("请选择已收盘日期")
    if args.provider == "public":
        if end != now.date():
            parser.error("公开收盘快照仅支持当天，历史日期请选择 BaoStock")
        if now.hour < 15 and not args.prewarm:
            print("REVIEW 北京时间 15:00 后开始核对收盘行情", flush=True)
            raise SystemExit(75)
    state = SERVER.parent / "tmp"
    state.mkdir(exist_ok=True)
    with (state / "market-review.lock").open("a") as lock, tempfile.TemporaryDirectory(prefix="market-review-") as folder:
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            print("REVIEW 正在等待其他复盘任务释放归档锁", flush=True)
            fcntl.flock(lock, fcntl.LOCK_EX)
        work = Path(folder)
        importer = work / "import-review"
        env = os.environ.copy()
        env.setdefault("GOCACHE", str(state / "go-build"))
        if installed := os.environ.get("REVIEW_IMPORTER"):
            importer = Path(installed).resolve()
        else:
            subprocess.run(["go", "build", "-o", str(importer), "./cmd/import-review"], cwd=SERVER, env=env, check=True)
        try:
            print("REVIEW 核对交易日和已归档数据", flush=True)
            from review_close import sync_public, sync_baostock
            if args.provider == "public":
                completed = sync_public(work, state, end, importer, warm_only=args.prewarm,
                                        core_only=args.core_only, scheduled=args.scheduled)
                if completed is False:
                    raise SystemExit(76)
            else:
                sync_baostock(work, state, end, importer, core_only=args.core_only)
        except QuotesNotPublished as error:
            print("REVIEW " + str(error), flush=True)
            raise SystemExit(75)
        except ValueError as error:
            print("REVIEW " + str(error), flush=True)
            raise


if __name__ == "__main__":
    main()
