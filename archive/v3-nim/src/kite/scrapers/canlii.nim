import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  CanLIIScraper* = ref object of BaseScraper

method baseUrl*(self: CanLIIScraper): string = "https://www.canlii.org"
method jurisdiction*(self: CanLIIScraper): string = "Canada"

proc newCanLIIScraper*(): CanLIIScraper = CanLIIScraper(initBaseScraper())

proc parseSearchHtml(self: CanLIIScraper; html: string; language = "en"): seq[CaseData] =
  var out: seq[CaseData] = @[]
  let resultRe = re"<div class=\"result\">([\s\S]*?)</div>"
  for rm in findAll(html, resultRe):
    let block = rm.group(0,1)
    let titleRe = re"<a[^>]*class=\"title\"[^>]*href=\"([^"]+)\"[^>]*>(.*?)</a>"
    let tm = titleRe.find(block)
    if tm.matchedLen == 0: continue
    var c = initCaseData()
    c.caseName = sanitizeText(block.substr(tm.group(0,2).a, tm.group(0,2).b))
    let href = block.substr(tm.group(0,1).a, tm.group(0,1).b)
    c.url = if href.startsWith("http"): href else: self.baseUrl() & href

    # meta: court and date
    let metaRe = re"<div class=\"resultmeta\">([\s\S]*?)</div>"
    let mm = metaRe.find(block)
    if mm.matchedLen > 0:
      let meta = sanitizeText(block.substr(mm.group(0,1).a, mm.group(0,1).b))
      let cd = re"([^,]+),\s*(\d{4}-\d{2}-\d{2})"
      let cdm = cd.find(meta)
      if cdm.matchedLen > 0:
        c.court = normalizeCourtName(meta.substr(cdm.group(0,1).a, cdm.group(0,1).b))
        try:
          c.date = parse("yyyy-MM-dd", meta.substr(cdm.group(0,2).a, cdm.group(0,2).b))
        except: discard
      # citations
      for cm in findAll(meta, re"\d{4}\s+[A-Z]+\s+\d+"):
        c.citations.add sanitizeText(meta.substr(cm.group(0,0).a, cm.group(0,0).b))

    # summary
    let sumRe = re"<div class=\"summary\">([\s\S]*?)</div>"
    let sm = sumRe.find(block)
    if sm.matchedLen > 0:
      c.summary = sanitizeText(block.substr(sm.group(0,1).a, sm.group(0,1).b))

    # caseId from URL
    let idm = re"/([^/]+)\.html$".find(c.url)
    if idm.matchedLen > 0:
      c.caseId = c.url.substr(idm.group(0,1).a, idm.group(0,1).b)

    c.jurisdiction = self.jurisdiction()
    out.add c
  return out

proc searchCases*(self: CanLIIScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100; language = "en"): seq[CaseData] =
  var params: seq[(string,string)] = @[("resultCount", $(min(limit, 200))), ("sort","decisionDateDesc")]
  if query.len > 0: params.add ("text", query)
  if startDate.len > 0: params.add ("dateFrom", startDate)
  if endDate.len > 0: params.add ("dateTo", endDate)
  if court.len > 0: params.add ("tribunal", court)
  let url = self.baseUrl() & "/" & language & "/search/"
  let resp = self.makeRequest(url, params=params)
  result = self.parseSearchHtml(resp.body, language)
  if result.len > limit: result.setLen(limit)

proc getCaseById*(self: CanLIIScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url: string
  if caseId.startsWith("http"):
    url = caseId
  elif '/' in caseId:
    url = self.baseUrl() & "/en/" & caseId & ".html"
  else:
    let cs = self.searchCases("citation:\"" & caseId & "\"", limit=1)
    if cs.len > 0: return cs[0]
    return initCaseData()
  let resp = self.makeRequest(url)
  var c = initCaseData()
  c.url = url
  let titleRe = re"<title>(.*?)</title>"; let mm = titleRe.find(resp.body)
  if mm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(mm.group(0,1).a, mm.group(0,1).b))
  # court
  let courtRe = re"<span class=\"court\">([\s\S]*?)</span>"; let cm = courtRe.find(resp.body)
  if cm.matchedLen > 0:
    c.court = normalizeCourtName(sanitizeText(resp.body.substr(cm.group(0,1).a, cm.group(0,1).b)))
  # date
  let dateRe = re"<span class=\"date\">([\s\S]*?)</span>"; let dm = dateRe.find(resp.body)
  if dm.matchedLen > 0:
    let dt = sanitizeText(resp.body.substr(dm.group(0,1).a, dm.group(0,1).b))
    let dmatch = re"(\d{4}-\d{2}-\d{2})".find(dt)
    if dmatch.matchedLen > 0:
      try: c.date = parse("yyyy-MM-dd", dt.substr(dmatch.group(0,1).a, dmatch.group(0,1).b)) except: discard
  # citation
  let citRe = re"<span class=\"citation\">([\s\S]*?)</span>"; let cim = citRe.find(resp.body)
  if cim.matchedLen > 0:
    let ct = sanitizeText(resp.body.substr(cim.group(0,1).a, cim.group(0,1).b))
    if ct.len > 0: c.citations.add ct
  # full text heuristic
  let contRe = re"<div class=\"(documentcontent|document|content)\">([\s\S]*?)</div>"
  let contm = contRe.find(resp.body)
  if contm.matchedLen > 0:
    c.fullText = sanitizeText(resp.body.substr(contm.group(0,2).a, contm.group(0,2).b))
  # judges heuristic
  if c.fullText.len > 0:
    for jm in findAll(c.fullText, re"(?:Justice|Judge|J\.)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)"):
      let name = c.fullText.substr(jm.group(0,1).a, jm.group(0,1).b)
      if c.judges.len < 5 and not (name in c.judges): c.judges.add name
  # id
  let idm = re"/([^/]+)\.html$".find(url)
  if idm.matchedLen > 0: c.caseId = url.substr(idm.group(0,1).a, idm.group(0,1).b)
  c.jurisdiction = self.jurisdiction()
  return c
