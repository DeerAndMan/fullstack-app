"""Available quote sources for the review archive."""
import importlib.util

PRICE_FIELDS = ("open", "high", "low", "close")
SOURCES = {
    "public": ("腾讯／新浪收盘行情", "https://stock.qq.com/"),
    "baostock": ("BaoStock 免费日线", "https://baostock.com/mainContent?file=DailyUpdates.md"),
}


class QuotesNotPublished(ValueError):
    """The requested daily market data has not been published."""


def options():
    installed = importlib.util.find_spec("baostock") is not None
    return [dict(id=key, label=label, reason="") for key, (label, _) in SOURCES.items()
            if installed]
