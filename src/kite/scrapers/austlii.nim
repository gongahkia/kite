import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  AustLIIScraper* = ref object of BaseScraper

method baseUrl*(self: AustLIIScraper): string = "https://www.austlii.edu.au"
method jurisdiction*(self: AustLIIScraper): string = "Australia"

proc newAustLIIScraper*(): AustLIIScraper = AustLIIScraper(initBaseScraper())

proc searchCases*(self: AustLIIScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100): seq[CaseData] =
  var params: seq[(string,string)] = @[]
  if query.len > 0: params.add ("query", query)
  if startDate.len > 0: params.add ("dfrom", startDate)
  if endDate.len > 0: params.add ("dto", endDate)
  if court.len > 0: params.add ("court", court)
  params.add ("results", $(min(limit, 200)))
  let url = self.baseUrl() & "/cgi-bin/sinodisp/au/cases/cth/HCA/"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  let linkRe = re"<a[^>]+href=\"(/au/cases/[^\"]+)\"[^>]*>(.*?)</a>"
  for m in findAll(resp.body, linkRe):
    var c = initCaseData()
    c.caseName = sanitizeText(m.group(0,2))
    let href = m.group(0,1)
    c.url = if href.startsWith("http"): href else: self.baseUrl() & href
    let idRe = re"/au/cases/([^/]+/[^/]+/[^/]+/\d+/\d+)"
    let idm = idRe.find(c.url)
    if idm.matchedLen > 0: c.caseId = c.url.substr(idm.group(0,1).a, idm.group(0,1).b)
    c.jurisdiction = self.jurisdiction()
    out.add c
    if out.len >= limit: break
  return out

proc getCaseById*(self: AustLIIScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url = caseId
  if not url.startsWith("http"):
    if url.startsWith("au/cases/"): url = self.baseUrl() & "/" & url & ".html"
  let resp = self.makeRequest(url)
  var c = initCaseData()
  c.url = url
  let titleRe = re"<title>(.*?)</title>"; let mm = titleRe.find(resp.body)
  if mm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(mm.group(0,1).a, mm.group(0,1).b))

  # court
  let courtPat = re"(High Court of Australia|Federal Court of Australia|NSW Court of Appeal|Victorian Court of Appeal)"
  let cm = courtPat.find(resp.body)
  if cm.matchedLen > 0: c.court = normalizeCourtName(sanitizeText(resp.body.substr(cm.group(0,1).a, cm.group(0,1).b)))

  # date patterns
  let datePatterns = [re"(\d{1,2}\s+(?:January|February|March|April|May|June|July|August|September|October|November|December)\s+\d{4})", re"(\d{4}-\d{2}-\d{2})", re"(\d{1,2}/\d{1,2}/\d{4})"]
  for dp in datePatterns:
    let dm = dp.find(resp.body)
    if dm.matchedLen > 0:
      let dateStr = resp.body.substr(dm.group(0,1).a, dm.group(0,1).b)
      try:
        if dateStr.contains('-'): c.date = parse("yyyy-MM-dd", dateStr)
        elif dateStr.contains('/'): c.date = parse("dd/MM/yyyy", dateStr)
        else: c.date = parse("d MMMM yyyy", dateStr)
      except: discard
      break

  # citations
  for cm2 in findAll(resp.body, re"\[(\d{4})\]\s+HCA\s+(\d+)"):
    c.citations.add "[" & resp.body.substr(cm2.group(0,1).a, cm2.group(0,1).b) & "] HCA " & resp.body.substr(cm2.group(0,2).a, cm2.group(0,2).b)

  # full text heuristic
  let contentRe = re"<div[^>]*(?:class=\"(judgment|content)\"|id=\"main\")[^>]*>([\s\S]*?)</div>"
  let contm = contentRe.find(resp.body)
  if contm.matchedLen > 0: c.fullText = sanitizeText(resp.body.substr(contm.group(0,2).a, contm.group(0,2).b))

  # judges
  if c.fullText.len > 0:
    for jm in findAll(c.fullText, re"(?:Justice|J\.)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)"):
      let name = c.fullText.substr(jm.group(0,1).a, jm.group(0,1).b).replace(" J.", "").replace(" J", "")
      if c.judges.len < 5 and not (name in c.judges): c.judges.add name

  # id
  let idm = re"/au/cases/([^/]+/[^/]+/[^/]+/\d+/\d+)".find(url)
  if idm.matchedLen > 0: c.caseId = url.substr(idm.group(0,1).a, idm.group(0,1).b)
  c.jurisdiction = self.jurisdiction()
  return c
