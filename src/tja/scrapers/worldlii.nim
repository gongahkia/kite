import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  WorldLIIScraper* = ref object of BaseScraper

method baseUrl*(self: WorldLIIScraper): string = "https://www.worldlii.org"
method jurisdiction*(self: WorldLIIScraper): string = "International"

proc newWorldLIIScraper*(): WorldLIIScraper = WorldLIIScraper(initBaseScraper())

proc searchCases*(self: WorldLIIScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100; jurisdiction = ""): seq[CaseData] =
  var params: seq[(string,string)] = @[("method","boolean"), ("results", $(min(limit, 200)))]
  if query.len > 0: params.add ("query", query)
  if startDate.len > 0: params.add ("dfrom", startDate)
  if endDate.len > 0: params.add ("dto", endDate)
  if court.len > 0: params.add ("court", court)
  if jurisdiction.len > 0: params.add ("db", jurisdiction)
  let url = self.baseUrl() & "/cgi-bin/sinodisp/int/cases/"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"(/int/cases/[^\"]+)\"[^>]*>(.*?)</a>"):
    var c = initCaseData(); c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1); c.url = if href.startsWith("http"): href else: self.baseUrl() & href
    let idm = re"/int/cases/([^/]+/\d+/\d+)".find(c.url); if idm.matchedLen > 0: c.caseId = "int/cases/" & c.url.substr(idm.group(0,1).a, idm.group(0,1).b)
    c.jurisdiction = self.jurisdiction(); out.add c; if out.len >= limit: break
  return out

proc getCaseById*(self: WorldLIIScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url = caseId
  if not url.startsWith("http"):
    if url.startsWith("int/cases/"): url = self.baseUrl() & "/" & url & ".html" else: return initCaseData()
  let resp = self.makeRequest(url)
  var c = initCaseData(); c.url = url
  let tm = re"<title>(.*?)</title>".find(resp.body); if tm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(tm.group(0,1).a, tm.group(0,1).b))
  c.jurisdiction = self.jurisdiction(); return c
