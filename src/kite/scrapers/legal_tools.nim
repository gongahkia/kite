import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText


type
  LegalToolsScraper* = ref object of BaseScraper

method baseUrl*(self: LegalToolsScraper): string = "https://www.legal-tools.org"
method jurisdiction*(self: LegalToolsScraper): string = "International Criminal Law"

proc newLegalToolsScraper*(): LegalToolsScraper = LegalToolsScraper(initBaseScraper())

proc searchCases*(self: LegalToolsScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100; documentType = ""; language = ""): seq[CaseData] =
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("q", query)
  if startDate.len > 0: params.add ("start_date", startDate)
  if endDate.len > 0: params.add ("end_date", endDate)
  if court.len > 0: params.add ("court", court)
  if documentType.len > 0: params.add ("doc_type", documentType)
  if language.len > 0: params.add ("lang", language)
  params.add ("limit", $(min(limit, 200)))
  let url = self.baseUrl() & "/doc/search"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"(/doc/\d+/)\"[^>]*>(.*?)</a>"):
    var c = initCaseData(); c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1); c.url = self.baseUrl() & href
    let idm = re"/doc/(\d+)/".find(c.url); if idm.matchedLen > 0: c.caseId = c.url.substr(idm.group(0,1).a, idm.group(0,1).b)
    c.jurisdiction = self.jurisdiction(); out.add c; if out.len >= limit: break
  return out

proc getCaseById*(self: LegalToolsScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url = caseId
  if not url.startsWith("http"):
    if caseId.allCharsInSet({'0'..'9'}): url = self.baseUrl() & "/doc/" & caseId & "/" else: return initCaseData()
  let resp = self.makeRequest(url)
  var c = initCaseData(); c.url = url
  let tm = re"<title>(.*?)</title>".find(resp.body); if tm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(tm.group(0,1).a, tm.group(0,1).b))
  c.jurisdiction = self.jurisdiction(); return c
