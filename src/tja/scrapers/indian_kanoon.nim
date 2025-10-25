import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  IndianKanoonScraper* = ref object of BaseScraper

method baseUrl*(self: IndianKanoonScraper): string = "https://indiankanoon.org"
method jurisdiction*(self: IndianKanoonScraper): string = "India"

proc newIndianKanoonScraper*(): IndianKanoonScraper = IndianKanoonScraper(initBaseScraper())

proc parseResult(self: IndianKanoonScraper; block: string): CaseData =
  var c = initCaseData()
  let titleRe = re"<a[^>]*class=\"result_title\"[^>]*href=\"([^"]+)\"[^>]*>(.*?)</a>"
  let tm = titleRe.find(block)
  if tm.matchedLen == 0: return c
  let href = block.substr(tm.group(0,1).a, tm.group(0,1).b)
  c.caseName = sanitizeText(block.substr(tm.group(0,2).a, tm.group(0,2).b))
  c.url = if href.startsWith("http"): href else: self.baseUrl() & href
  let idm = re"/doc/(\d+)/".find(c.url)
  if idm.matchedLen > 0: c.caseId = c.url.substr(idm.group(0,1).a, idm.group(0,1).b)
  # meta
  let metaRe = re"<div class=\"result_meta\">([\s\S]*?)</div>"
  let mm = metaRe.find(block)
  if mm.matchedLen > 0:
    let meta = sanitizeText(block.substr(mm.group(0,1).a, mm.group(0,1).b))
    let courtm = re"Court:\s*([^,\n]+)".find(meta, {reIgnoreCase})
    if courtm.matchedLen > 0:
      c.court = normalizeCourtName(meta.substr(courtm.group(0,1).a, courtm.group(0,1).b))
    let dm = re"(\d{1,2}-\d{1,2}-\d{4})".find(meta)
    if dm.matchedLen > 0:
      try: c.date = parse("dd-MM-yyyy", meta.substr(dm.group(0,1).a, dm.group(0,1).b)) except: discard
    for cm in findAll(meta, re"(\d{4})\s+(\d+)\s+(SCC|SCR|AIR)"):
      let y = meta.substr(cm.group(0,1).a, cm.group(0,1).b)
      let n = meta.substr(cm.group(0,2).a, cm.group(0,2).b)
      let rep = meta.substr(cm.group(0,3).a, cm.group(0,3).b)
      c.citations.add "(" & y & ") " & n & " " & rep
  # summary
  let sumRe = re"<div class=\"result_summary\">([\s\S]*?)</div>"
  let sm = sumRe.find(block)
  if sm.matchedLen > 0:
    c.summary = sanitizeText(block.substr(sm.group(0,1).a, sm.group(0,1).b))
  c.jurisdiction = self.jurisdiction()
  return c

proc searchCases*(self: IndianKanoonScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100): seq[CaseData] =
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("formInput", query)
  if startDate.len > 0: params.add ("from_date", startDate)
  if endDate.len > 0: params.add ("to_date", endDate)
  if court.len > 0: params.add ("court", court)
  params.add ("limit", $(min(limit, 200)))
  let url = self.baseUrl() & "/search/"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  for m in findAll(resp.body, re"<div class=\"result\">([\s\S]*?)</div>"):
    if out.len >= limit: break
    out.add self.parseResult(resp.body.substr(m.group(0,1).a, m.group(0,1).b))
  return out

proc getCaseById*(self: IndianKanoonScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url: string
  if caseId.startsWith("http"): url = caseId
  elif caseId.allCharsInSet({'0'..'9'}): url = self.baseUrl() & "/doc/" & caseId & "/"
  else:
    let r = self.searchCases(caseId, limit=1); if r.len > 0: return r[0] else: return initCaseData()
  let resp = self.makeRequest(url)
  var c = initCaseData(); c.url = url
  let titleRe = re"<title>(.*?)</title>"; let tm = titleRe.find(resp.body)
  if tm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(tm.group(0,1).a, tm.group(0,1).b))
  c.jurisdiction = self.jurisdiction()
  return c
