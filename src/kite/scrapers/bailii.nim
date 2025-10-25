import std/[re, strutils, times]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText, normalizeCourtName


type
  BAILIIScraper* = ref object of BaseScraper

method baseUrl*(self: BAILIIScraper): string = "https://www.bailii.org"
method jurisdiction*(self: BAILIIScraper): string = "United Kingdom"

proc newBAILIIScraper*(): BAILIIScraper = BAILIIScraper(initBaseScraper())

proc searchCases*(self: BAILIIScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100; jurisdiction = "uk"): seq[CaseData] =
  var params: seq[(string,string)] = @[("method","boolean"), ("results", $(min(limit, 200)))]
  if query.len > 0: params.add ("query", query)
  if startDate.len > 0: params.add ("date_from", startDate)
  if endDate.len > 0: params.add ("date_to", endDate)
  if court.len > 0: params.add ("court", court)
  if jurisdiction in ["uk","ie","ni","scot"]: params.add ("jurisdiction", jurisdiction)
  let url = self.baseUrl() & "/cgi-bin/markup.cgi"
  let resp = self.makeRequest(url, params=params)
  var out: seq[CaseData] = @[]
  let linkRe = re"<a[^>]+href=\"(/(uk|ie|ni|scot)/cases/[^\"]+)\"[^>]*>(.*?)</a>"
  for m in findAll(resp.body, linkRe):
    var c = initCaseData()
    c.caseName = sanitizeText(m.group(0,3))
    let href = m.group(0,1)
    c.url = if href.startsWith("http"): href else: self.baseUrl() & href
    let idRe = re"/(uk|ie|ni|scot)/cases/([^/]+/\d+/\d+)"
    let mm = idRe.find(c.url)
    if mm.matchedLen > 0:
      c.caseId = c.url.substr(mm.group(0,2).a, mm.group(0,2).b)
    if "/ie/cases/" in c.url: c.jurisdiction = "Ireland"
    elif "/ni/cases/" in c.url: c.jurisdiction = "Northern Ireland"
    elif "/scot/cases/" in c.url: c.jurisdiction = "Scotland"
    else: c.jurisdiction = self.jurisdiction()
    out.add c
    if out.len >= limit: break
  return out

proc getCaseById*(self: BAILIIScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url = caseId
  if not url.startsWith("http"):
    if url.startsWith("uk/cases/") or url.startsWith("ie/cases/") or url.startsWith("ni/cases/") or url.startsWith("scot/cases/"):
      url = self.baseUrl() & "/" & url & ".html"
  let resp = self.makeRequest(url)
  var c = initCaseData()
  c.url = url
  let titleRe = re"<title>(.*?)</title>"; let mm = titleRe.find(resp.body)
  if mm.matchedLen > 0: c.caseName = sanitizeText(resp.body.substr(mm.group(0,1).a, mm.group(0,1).b))

  # court patterns
  let courtPatterns = [re"(Supreme Court|Court of Appeal|High Court|Crown Court|Magistrates|Employment Tribunal)", re"(UKSC|EWCA|EWHC|UKUT|UKFTT)", re"(House of Lords|Privy Council)"]
  for cp in courtPatterns:
    let cm = cp.find(resp.body)
    if cm.matchedLen > 0:
      c.court = normalizeCourtName(sanitizeText(resp.body.substr(cm.group(0,1).a, cm.group(0,1).b)))
      break

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
  for cm2 in findAll(resp.body, re"\[(\d{4})\]\s+(UKSC|EWCA|EWHC|UKUT)\s+(\d+)"):
    let y = resp.body.substr(cm2.group(0,1).a, cm2.group(0,1).b)
    let courtCode = resp.body.substr(cm2.group(0,2).a, cm2.group(0,2).b)
    let num = resp.body.substr(cm2.group(0,3).a, cm2.group(0,3).b)
    c.citations.add "[" & y & "] " & courtCode & " " & num

  # full text
  let contentRe = re"<div[^>]*(?:class=\"(judgment|content)\"|id=\"main\")[^>]*>([\s\S]*?)</div>"
  let contm = contentRe.find(resp.body)
  if contm.matchedLen > 0: c.fullText = sanitizeText(resp.body.substr(contm.group(0,2).a, contm.group(0,2).b))

  # judges
  if c.fullText.len > 0:
    for jm in findAll(c.fullText, re"(?:Lord|Lady|Mr|Mrs|Ms)\s+Justice\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)|([A-Z][a-z]+\s+LJ)|([A-Z][a-z]+\s+J\.?)"):
      let seg = c.fullText.substr(jm.group(0,0).a, jm.group(0,0).b)
      var name = seg.replace(" LJ", "").replace(" J.", "").replace(" J", "")
      name = name.replace(re"^(?:Lord|Lady|Mr|Mrs|Ms)\s+Justice\s+", "")
      if name.len > 0 and c.judges.len < 5 and not (name in c.judges): c.judges.add name

  # id & jurisdiction from URL
  let idm = re"/(uk|ie|ni|scot)/cases/([^/]+/\d+/\d+)".find(url)
  if idm.matchedLen > 0:
    c.caseId = url.substr(idm.group(0,2).a, idm.group(0,2).b)
    let jur = url.substr(idm.group(0,1).a, idm.group(0,1).b)
    c.jurisdiction = (if jur == "ie": "Ireland" elif jur == "ni": "Northern Ireland" elif jur == "scot": "Scotland" else: self.jurisdiction())
  return c
