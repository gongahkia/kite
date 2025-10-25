import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  HKLIIScraper* = ref object of BaseScraper

method baseUrl*(self: HKLIIScraper): string = "https://www.hklii.hk"
method jurisdiction*(self: HKLIIScraper): string = "Hong Kong"

proc newHKLIIScraper*(): HKLIIScraper = HKLIIScraper(initBaseScraper())

proc searchCases*(self: HKLIIScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100; language = "en"): seq[CaseData] =
  var params: seq[(string,string)] = @[("method","boolean"), ("results", $(min(limit, 200)))]
  if query.len > 0: params.add ("query", query)
  if startDate.len > 0: params.add ("dfrom", startDate)
  if endDate.len > 0: params.add ("dto", endDate)
  if court.len > 0: params.add ("court", court)
  if language in ["en","zh"]: params.add ("language", language)
  let url = self.baseUrl() & "/cgi-bin/sinodisp/hk/cases/hkcfa/"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"(/hk/cases/[^\"]+)\"[^>]*>(.*?)</a>"):
    var c = initCaseData()
    c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1)
    c.url = if href.startsWith("http"): href else: self.baseUrl() & href
    let idm = re"/hk/cases/([^/]+/\d+/\d+)".find(c.url)
    if idm.matchedLen > 0: c.caseId = "hk/cases/" & c.url.substr(idm.group(0,1).a, idm.group(0,1).b)
    c.jurisdiction = self.jurisdiction()
    out.add c
    if out.len >= limit: break
  return out

proc getCaseById*(self: HKLIIScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url = caseId
  if not url.startsWith("http"):
    if url.startsWith("hk/cases/"): url = self.baseUrl() & "/" & url & ".html" else: return initCaseData()
  let resp = self.makeRequest(url)
  var c = initCaseData(); c.url = url; c.jurisdiction = self.jurisdiction()
  let titleRe = re"<title>(.*?)</title>"; let tm = titleRe.find(resp.body)
  if tm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(tm.group(0,1).a, tm.group(0,1).b))
  return c
