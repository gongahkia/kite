import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText


type
  FindLawScraper* = ref object of BaseScraper

method baseUrl*(self: FindLawScraper): string = "https://caselaw.findlaw.com"
method jurisdiction*(self: FindLawScraper): string = "United States"

proc newFindLawScraper*(): FindLawScraper = FindLawScraper(initBaseScraper())

proc scrapeCaseFromUrl(self: FindLawScraper; url: string): CaseData =
  let resp = self.makeRequest(url)
  var c = initCaseData()
  c.url = url
  let titleRe = re"<h1[^>]*>(.*?)</h1>|<title>(.*?)</title>"
  let tm = titleRe.find(resp.body)
  if tm.matchedLen > 0:
    let seg = if tm.group(0,1).a >= 0: tm.group(0,1) else: tm.group(0,2)
    c.caseName = sanitizeText(resp.body.substr(seg.a, seg.b))
  c.court = "FindLaw"
  c.jurisdiction = self.jurisdiction()
  let idm = re"/case/([^/]+)".find(url)
  if idm.matchedLen > 0:
    c.caseId = url.substr(idm.group(0,1).a, idm.group(0,1).b)
  # content
  let contentRe = re"<div[^>]*class=\"content\"[^>]*>([\s\S]*?)</div>|<main[^>]*>([\s\S]*?)</main>"
  let cm = contentRe.find(resp.body)
  if cm.matchedLen > 0:
    let seg2 = if cm.group(0,1).a >= 0: cm.group(0,1) else: cm.group(0,2)
    c.fullText = sanitizeText(resp.body.substr(seg2.a, seg2.b))
  return c

proc searchCases*(self: FindLawScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100): seq[CaseData] =
  var url = self.baseUrl() & "/us-supreme-court/"
  if court.len > 0 and court.toLowerAscii != "supreme":
    url = self.baseUrl() & "/state/"
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("q", query)
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  let linkRe = re"<a[^>]+href=\"(/case/[^\"]+)\"[^>]*>"
  for m in findAll(resp.body, linkRe):
    if out.len >= limit: break
    let href = m.group(0,1)
    let caseUrl = if href.startsWith("http"): href else: self.baseUrl() & href
    try:
      out.add self.scrapeCaseFromUrl(caseUrl)
    except: continue
  return out

proc getCaseById*(self: FindLawScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  let url = self.baseUrl() & "/case/" & caseId
  return self.scrapeCaseFromUrl(url)

proc extractFindlawIdFromUrl*(url: string): string =
  let idm = re"/case/([^/]+)".find(url)
  if idm.matchedLen > 0: return url.substr(idm.group(0,1).a, idm.group(0,1).b)
  ""
