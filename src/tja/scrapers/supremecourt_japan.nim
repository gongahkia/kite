import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText


type
  SupremeCourtJapanScraper* = ref object of BaseScraper

method baseUrl*(self: SupremeCourtJapanScraper): string = "https://www.courts.go.jp"
method jurisdiction*(self: SupremeCourtJapanScraper): string = "Japan"

proc newSupremeCourtJapanScraper*(): SupremeCourtJapanScraper = SupremeCourtJapanScraper(initBaseScraper())

proc searchCases*(self: SupremeCourtJapanScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100; language = "en"): seq[CaseData] =
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("q", query)
  if startDate.len > 0: params.add ("start_date", startDate)
  if endDate.len > 0: params.add ("end_date", endDate)
  params.add ("lang", language); params.add ("limit", $(min(limit, 200)))
  let url = self.baseUrl() & "/app/hanrei_en/search"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"(/app/hanrei_en/detail[^\"]+)\"[^>]*>(.*?)</a>"):
    var c = initCaseData(); c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1); c.url = self.baseUrl() & href
    let idm = re"id=([^&]+)".find(c.url); if idm.matchedLen > 0: c.caseId = c.url.substr(idm.group(0,1).a, idm.group(0,1).b)
    c.court = "Supreme Court of Japan"; c.jurisdiction = self.jurisdiction(); out.add c; if out.len >= limit: break
  return out

proc getCaseById*(self: SupremeCourtJapanScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  let res = self.searchCases(caseId, limit=1); if res.len > 0: return res[0]
  return initCaseData()
