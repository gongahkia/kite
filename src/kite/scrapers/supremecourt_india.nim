import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText


type
  SupremeCourtIndiaScraper* = ref object of BaseScraper

method baseUrl*(self: SupremeCourtIndiaScraper): string = "https://main.sci.gov.in"
method jurisdiction*(self: SupremeCourtIndiaScraper): string = "India"

proc newSupremeCourtIndiaScraper*(): SupremeCourtIndiaScraper = SupremeCourtIndiaScraper(initBaseScraper())

proc searchCases*(self: SupremeCourtIndiaScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100): seq[CaseData] =
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("search_text", query)
  if startDate.len > 0: params.add ("from_date", startDate)
  if endDate.len > 0: params.add ("to_date", endDate)
  params.add ("limit", $(min(limit, 200)))
  let url = self.baseUrl() & "/judgments"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"(/judgment/[^\"]+)\"[^>]*>(.*?)</a>"):
    var c = initCaseData(); c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1); c.url = self.baseUrl() & href
    c.court = "Supreme Court of India"; c.jurisdiction = self.jurisdiction()
    out.add c; if out.len >= limit: break
  return out

proc getCaseById*(self: SupremeCourtIndiaScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  let res = self.searchCases(caseId, limit=1); if res.len > 0: return res[0]
  return initCaseData()
