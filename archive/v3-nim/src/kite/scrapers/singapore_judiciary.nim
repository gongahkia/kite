import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  SingaporeJudiciaryScraper* = ref object of BaseScraper

method baseUrl*(self: SingaporeJudiciaryScraper): string = "https://www.judiciary.gov.sg"
method jurisdiction*(self: SingaporeJudiciaryScraper): string = "Singapore"

proc newSingaporeJudiciaryScraper*(): SingaporeJudiciaryScraper = SingaporeJudiciaryScraper(initBaseScraper())

proc parseResultLink(self: SingaporeJudiciaryScraper; html: string; href: string): CaseData =
  var c = initCaseData()
  c.caseName = sanitizeText(html)
  var url = href
  if not url.startsWith("http"): url = self.baseUrl() & url
  c.url = url
  let citm = re"\[(\d{4})\]\s+(SGCA|SGHC|SGFC|SGMC)\s+(\d+)".find(c.caseName)
  if citm.matchedLen > 0:
    let y = c.caseName.substr(citm.group(0,1).a, citm.group(0,1).b)
    let code = c.caseName.substr(citm.group(0,2).a, citm.group(0,2).b)
    let num = c.caseName.substr(citm.group(0,3).a, citm.group(0,3).b)
    c.caseId = "[" & y & "] " & code & " " & num
  c.jurisdiction = self.jurisdiction()
  return c

proc searchCases*(self: SingaporeJudiciaryScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100): seq[CaseData] =
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("searchText", query)
  if startDate.len > 0: params.add ("startDate", startDate)
  if endDate.len > 0: params.add ("endDate", endDate)
  if court.len > 0: params.add ("court", court)
  params.add ("limit", $(min(limit, 200)))
  let url = self.baseUrl() & "/judgment-search"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  let linkRe = re"<a[^>]+href=\"(/judgment/[^\"]+)\"[^>]*>([\s\S]*?)</a>"
  for m in findAll(resp.body, linkRe):
    if out.len >= limit: break
    out.add self.parseResultLink(m.group(0,2), m.group(0,1))
  return out

proc getCaseById*(self: SingaporeJudiciaryScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  # If looks like citation, search
  if caseId.startsWith("[") and "]" in caseId:
    let r = self.searchCases(caseId, limit=1)
    if r.len > 0: return r[0]
    return initCaseData()
  # Otherwise, treat as URL
  var url = caseId
  if not url.startsWith("http"): return initCaseData()
  let resp = self.makeRequest(url)
  var c = initCaseData()
  c.url = url
  let titleRe = re"<title>(.*?)</title>"; let tm = titleRe.find(resp.body)
  if tm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(tm.group(0,1).a, tm.group(0,1).b))
  c.jurisdiction = self.jurisdiction()
  return c
