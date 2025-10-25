import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  CuriaEuropaScraper* = ref object of BaseScraper

method baseUrl*(self: CuriaEuropaScraper): string = "https://curia.europa.eu"
method jurisdiction*(self: CuriaEuropaScraper): string = "European Union"

proc newCuriaEuropaScraper*(): CuriaEuropaScraper = CuriaEuropaScraper(initBaseScraper())

proc searchCases*(self: CuriaEuropaScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100; language = "en"): seq[CaseData] =
  var params: seq[(string,string)] = @[("page","0"), ("size", $(min(limit, 100))), ("lang", language)]
  if query.len > 0: params.add ("text", query)
  if startDate.len > 0: params.add ("DD_date_start", startDate)
  if endDate.len > 0: params.add ("DD_date_end", endDate)
  if court.len > 0: params.add ("court", court)
  let url = self.baseUrl() & "/juris/liste.jsf"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"(/juris/document/[^\"]+)\"[^>]*>(.*?)</a>"):
    var c = initCaseData(); c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1); c.url = if href.startsWith("http"): href else: self.baseUrl() & href
    let idm = re"([CT]-\d+/\d+)".find(c.caseName)
    if idm.matchedLen > 0: c.caseId = c.caseName.substr(idm.group(0,1).a, idm.group(0,1).b)
    c.jurisdiction = self.jurisdiction(); out.add c
    if out.len >= limit: break
  return out

proc getCaseById*(self: CuriaEuropaScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  if caseId.startsWith("http"):
    let resp = self.makeRequest(caseId); var c = initCaseData(); c.url = caseId
    let tm = re"<title>(.*?)</title>".find(resp.body); if tm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(tm.group(0,1).a, tm.group(0,1).b))
    c.jurisdiction = self.jurisdiction(); return c
  let res = self.searchCases(caseId, limit=1); if res.len > 0: return res[0]
  return initCaseData()
