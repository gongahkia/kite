import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  LegifranceScraper* = ref object of BaseScraper

method baseUrl*(self: LegifranceScraper): string = "https://www.legifrance.gouv.fr"
method jurisdiction*(self: LegifranceScraper): string = "France"

proc newLegifranceScraper*(): LegifranceScraper = LegifranceScraper(initBaseScraper())

proc searchCases*(self: LegifranceScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100; chamber = ""): seq[CaseData] =
  var params: seq[(string,string)] = @[("typePagination","defaut"), ("sortValue","SIGNATURE_DATE_DESC"), ("size", $(min(limit, 100)))]
  if query.len > 0: params.add ("recherche", query)
  if startDate.len > 0: params.add ("dateSignatureDebut", startDate)
  if endDate.len > 0: params.add ("dateSignatureFin", endDate)
  if court.len > 0: params.add ("juridiction", court)
  if chamber.len > 0: params.add ("chambre", chamber)
  let url = self.baseUrl() & "/search/juri"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<a[^>]+href=\"(/juri/[^\"]+)\"[^>]*>(.*?)</a>"):
    var c = initCaseData()
    c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1)
    c.url = if href.startsWith("http"): href else: self.baseUrl() & href
    let idm = re"/id/([A-Z]+\d+)".find(c.url)
    if idm.matchedLen > 0: c.caseId = c.url.substr(idm.group(0,1).a, idm.group(0,1).b)
    c.jurisdiction = self.jurisdiction()
    out.add c
    if out.len >= limit: break
  return out

proc getCaseById*(self: LegifranceScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url = caseId
  if not url.startsWith("http"):
    if caseId.startsWith("CETATEXT") or caseId.startsWith("JURITEXT"): url = self.baseUrl() & "/juri/id/" & caseId else: return initCaseData()
  let resp = self.makeRequest(url)
  var c = initCaseData(); c.url = url; c.jurisdiction = self.jurisdiction()
  let titleRe = re"<title>(.*?)</title>"; let tm = titleRe.find(resp.body)
  if tm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(tm.group(0,1).a, tm.group(0,1).b))
  let idm = re"/id/([A-Z]+\d+)".find(url)
  if idm.matchedLen > 0: c.caseId = url.substr(idm.group(0,1).a, idm.group(0,1).b)
  return c
