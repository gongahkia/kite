import std/[json, times, strutils]
import chronicles
from ../../tja/base_scraper import BaseScraper, initBaseScraper, makeRequest
from ../../tja/data_models import CaseData, initCaseData
from ../../tja/helpers import sanitizeText


type
  CourtListenerScraper* = ref object of BaseScraper

method baseUrl*(self: CourtListenerScraper): string = "https://www.courtlistener.com"
method jurisdiction*(self: CourtListenerScraper): string = "United States"

proc newCourtListenerScraper*(): CourtListenerScraper =
  result = CourtListenerScraper(initBaseScraper())

proc parseSearchItem(item: JsonNode; baseUrl: string): CaseData =
  var c = initCaseData()
  c.caseName = sanitizeText(item.getOrDefault("caseName").getStr)
  c.caseId = $item.getOrDefault("id").getInt
  if item.hasKey("dateFiled"):
    try:
      c.date = parse("yyyy-MM-dd", item["dateFiled"].getStr)
    except: discard
  c.court = sanitizeText(item.getOrDefault("court").getStr)
  if item.hasKey("absolute_url"):
    c.url = baseUrl & item["absolute_url"].getStr
  if item.hasKey("citation"):
    c.citations = @[item["citation"].getStr]
  if item.hasKey("judge"):
    c.judges = @[sanitizeText(item["judge"].getStr)]
  c.summary = sanitizeText(item.getOrDefault("snippet").getStr)
  c.jurisdiction = "United States"
  c.caseType = sanitizeText(item.getOrDefault("status").getStr)
  return c

proc searchCases*(self: CourtListenerScraper; query: string; startDate = ""; endDate = ""; court = ""; limit = 100): seq[CaseData] =
  var params: seq[(string,string)] = @[("format","json"), ("order_by","-date_filed")]
  if query.len > 0: params.add ("q", query)
  if startDate.len > 0: params.add ("filed_after", startDate)
  if endDate.len > 0: params.add ("filed_before", endDate)
  if court.len > 0: params.add ("court", court)
  if limit > 0: params.add ("count", $(min(limit, 1000)))
  let url = self.baseUrl() & "/api/rest/v3/search/"
  let resp = self.makeRequest(url, params=params)
  let data = parseJson(resp.body)
  let results = data.getOrDefault("results")
  if results.kind == JArray:
    for item in results.items:
      try:
        result.add parseSearchItem(item, self.baseUrl())
      except: continue
  return result

proc getCaseById*(self: CourtListenerScraper; caseId: string): CaseData =
  if caseId.len == 0: return initCaseData()
  var url = self.baseUrl() & "/api/rest/v3/opinions/" & caseId & "/"
  try:
    let resp = self.makeRequest(url, params = @[("format","json")])
    let data = parseJson(resp.body)
    # cluster lookup
    if data.hasKey("cluster"):
      let clusterUrl = data["cluster"].getStr
      let cresp = self.makeRequest(clusterUrl, params = @[("format","json")])
      let cdata = parseJson(cresp.body)
      var c = initCaseData()
      c.caseName = sanitizeText(cdata.getOrDefault("case_name").getStr)
      if cdata.hasKey("date_filed"):
        try: c.date = parse("yyyy-MM-dd", cdata["date_filed"].getStr) except: discard
      c.court = sanitizeText(cdata.getOrDefault("court").getStr)
      c.url = self.baseUrl() & cdata.getOrDefault("absolute_url").getStr
      c.caseId = $data.getOrDefault("id").getInt
      c.jurisdiction = self.jurisdiction()
      return c
  except: discard
  # try cluster endpoint directly
  url = self.baseUrl() & "/api/rest/v3/clusters/" & caseId & "/"
  try:
    let resp = self.makeRequest(url, params = @[("format","json")])
    let data = parseJson(resp.body)
    var c = initCaseData()
    c.caseName = sanitizeText(data.getOrDefault("case_name").getStr)
    if data.hasKey("date_filed"):
      try: c.date = parse("yyyy-MM-dd", data["date_filed"].getStr) except: discard
    c.court = sanitizeText(data.getOrDefault("court").getStr)
    c.url = self.baseUrl() & data.getOrDefault("absolute_url").getStr
    c.caseId = $data.getOrDefault("id").getInt
    c.jurisdiction = self.jurisdiction()
    return c
  except: discard
  return initCaseData()
