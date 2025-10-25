import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  GermanLawArchiveScraper* = ref object of BaseScraper

method baseUrl*(self: GermanLawArchiveScraper): string = "https://germanlawarchive.iuscomp.org"
method jurisdiction*(self: GermanLawArchiveScraper): string = "Germany"

proc newGermanLawArchiveScraper*(): GermanLawArchiveScraper = GermanLawArchiveScraper(initBaseScraper())

proc searchCases*(self: GermanLawArchiveScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100): seq[CaseData] =
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("q", query)
  if startDate.len > 0: params.add ("start_date", startDate)
  if endDate.len > 0: params.add ("end_date", endDate)
  if court.len > 0: params.add ("court", court)
  params.add ("limit", $(min(limit, 200)))
  let url = self.baseUrl() & "/search"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"(/cases/[^\"]+)\"[^>]*>(.*?)</a>"):
    var c = initCaseData()
    c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1)
    c.url = if href.startsWith("http"): href else: self.baseUrl() & href
    c.jurisdiction = self.jurisdiction()
    out.add c
    if out.len >= limit: break
  return out

proc getCaseById*(self: GermanLawArchiveScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  let res = self.searchCases(caseId, limit=1)
  if res.len > 0: return res[0]
  return initCaseData()
