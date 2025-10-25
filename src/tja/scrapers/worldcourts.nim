import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  WorldCourtsScraper* = ref object of BaseScraper

method baseUrl*(self: WorldCourtsScraper): string = "https://www.worldcourts.com"
method jurisdiction*(self: WorldCourtsScraper): string = "International"

proc newWorldCourtsScraper*(): WorldCourtsScraper = WorldCourtsScraper(initBaseScraper())

proc searchCases*(self: WorldCourtsScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100): seq[CaseData] =
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("q", query)
  if startDate.len > 0: params.add ("start_date", startDate)
  if endDate.len > 0: params.add ("end_date", endDate)
  if court.len > 0: params.add ("court", court)
  params.add ("limit", $(min(limit, 200)))
  let url = self.baseUrl() & "/search"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"/(inst|cases)/([^\"]+)\"[^>]*>(.*?)</a>"):
    var c = initCaseData(); c.caseName = sanitizeText(m.group(0,3))
    let seg = m.group(0,1) & "/" & m.group(0,2); c.url = self.baseUrl() & "/" & seg
    c.caseId = seg; c.jurisdiction = self.jurisdiction(); out.add c; if out.len >= limit: break
  return out

proc getCaseById*(self: WorldCourtsScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url = caseId; if not url.startsWith("http"): url = self.baseUrl() & "/" & url
  let resp = self.makeRequest(url)
  var c = initCaseData(); c.url = url
  let tm = re"<title>(.*?)</title>".find(resp.body); if tm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(tm.group(0,1).a, tm.group(0,1).b))
  c.jurisdiction = self.jurisdiction(); return c
