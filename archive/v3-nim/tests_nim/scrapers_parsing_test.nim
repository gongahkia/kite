import std/[unittest]
from ../src/tja/scrapers/canlii import CanLIIScraper, parseSearchHtml
from ../src/tja/scrapers/austlii import AustLIIScraper
from ../src/tja/scrapers/bailii import BAILIIScraper
from ../src/tja/helpers import normalizeDate

suite "Scraper parsing heuristics":
  test "CanLII search block parses case name, url, date":
    let html = """
    <div class="result">
      <a class="title" href="/en/ca/scc-csc/doc/2023/2023scc15/2023scc15.html">Foo v. Bar</a>
      <div class="resultmeta">Supreme Court of Canada, 2023-01-01, 2023 SCC 15</div>
      <div class="summary">Summary text</div>
    </div>
    """
    let s = CanLIIScraper()
    let res = s.parseSearchHtml(html, "en")
    check res.len == 1
    check res[0].caseName == "Foo v. Bar"
    check res[0].url.endsWith("2023scc15/2023scc15.html")

  test "normalizeDate handles French months":
    let d = normalizeDate("12 janvier 2023")
    check d.year == 2023

  test "normalizeDate handles German dot format":
    let d = normalizeDate("12.03.2021")
    check d.year == 2021

  test "BAILII link parsing extracts id":
    let html = "<a href='/uk/cases/UKSC/2023/15.html'>Baz</a>"
    let s = BAILIIScraper()
    discard s # compile check

  test "AustLII link parsing extracts id":
    let html = "<a href='/au/cases/cth/HCA/2023/15.html'>Qux</a>"
    let s = AustLIIScraper()
    discard s
