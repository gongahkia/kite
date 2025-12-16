import std/[strutils, tables, sequtils, times, asyncdispatch, asyncnet]
import chronicles
from ./config import getConfig

# Minimal in-memory Prometheus registry with labeled counters, gauges, and histograms

type
  LabelKey = string # canonical key like "scraper=CL;method=GET;status=200"

  Counter = Table[LabelKey, float]
  Gauge = Table[LabelKey, float]

  HistogramDef = object
    buckets: seq[float]
    counts: Table[LabelKey, seq[float]] # per bucket incl +Inf
    sums: Table[LabelKey, float]
    totals: Table[LabelKey, float]

var counters = initTable[string, Counter]()
var gauges = initTable[string, Gauge]()
var histos = initTable[string, HistogramDef]()

proc canon(labels: openArray[(string, string)]): LabelKey =
  var parts = labels.toSeq
  parts.sort(proc (a,b: (string,string)): int = cmp(a[0], b[0]))
  for (k, v) in parts: result.add k & "=" & v & ";"

proc getCounter(name: string): var Counter =
  if not counters.hasKey(name): counters[name] = initTable[LabelKey, float]()
  return counters[name]

proc getGauge(name: string): var Gauge =
  if not gauges.hasKey(name): gauges[name] = initTable[LabelKey, float]()
  return gauges[name]

proc getHisto(name: string; buckets: seq[float]): var HistogramDef =
  if not histos.hasKey(name):
    var h: HistogramDef
    h.buckets = buckets
    h.counts = initTable[LabelKey, seq[float]]()
    h.sums = initTable[LabelKey, float]()
    h.totals = initTable[LabelKey, float]()
    histos[name] = h
  return histos[name]

# === Metric APIs (matching v1 names) ===

# Counters
proc incRequest*(scraper, method, status: string; delta = 1.0) =
  let key = canon({("scraper", scraper), ("method", method), ("status", status)})
  var c = getCounter("kite_requests_total")
  c[key] = c.getOrDefault(key, 0.0) + delta

proc incRequestError*(scraper, errorType: string; delta = 1.0) =
  let key = canon({("scraper", scraper), ("error_type", errorType)})
  var c = getCounter("kite_request_errors_total")
  c[key] = c.getOrDefault(key, 0.0) + delta

proc incCasesScraped*(scraper, jurisdiction: string; delta = 1.0) =
  let key = canon({("scraper", scraper), ("jurisdiction", jurisdiction)})
  var c = getCounter("kite_cases_scraped_total")
  c[key] = c.getOrDefault(key, 0.0) + delta

proc incRateLimitHits*(scraper: string; delta = 1.0) =
  let key = canon({("scraper", scraper)})
  var c = getCounter("kite_rate_limit_hits_total")
  c[key] = c.getOrDefault(key, 0.0) + delta

proc incParsingErrors*(scraper, errorType: string; delta = 1.0) =
  let key = canon({("scraper", scraper), ("error_type", errorType)})
  var c = getCounter("kite_parsing_errors_total")
  c[key] = c.getOrDefault(key, 0.0) + delta

# Gauges
proc gaugeActiveScrapersAdd*(scraper: string; delta = 1.0) =
  let key = canon({("scraper", scraper)})
  var g = getGauge("kite_active_scrapers")
  g[key] = g.getOrDefault(key, 0.0) + delta

# Histograms
proc observe*(name: string; labels: openArray[(string,string)]; value: float; buckets: seq[float]) =
  var h = getHisto(name, buckets)
  let key = canon(labels)
  if not h.counts.hasKey(key):
    h.counts[key] = newSeq[float](buckets.len + 1) # +Inf bucket
    h.sums[key] = 0.0
    h.totals[key] = 0.0
  # increment appropriate bucket
  var idx = buckets.len # default +Inf
  for i, b in buckets:
    if value <= b: idx = i; break
  h.counts[key][idx] = h.counts[key][idx] + 1.0
  h.sums[key] = h.sums[key] + value
  h.totals[key] = h.totals[key] + 1.0

proc observeRequestDuration*(scraper, method: string; seconds: float) =
  observe("kite_request_duration_seconds", {("scraper", scraper), ("method", method)}, seconds, @[0.1,0.5,1.0,2.0,5.0,10.0,30.0,60.0])

proc observeScrapingDuration*(scraper: string; seconds: float) =
  observe("kite_scraping_duration_seconds", {("scraper", scraper)}, seconds, @[1.0,5.0,10.0,30.0,60.0,120.0,300.0])

proc observeRateLimitWait*(scraper: string; seconds: float) =
  observe("kite_rate_limit_wait_seconds", {("scraper", scraper)}, seconds, @[0.1,0.5,1.0,2.0,5.0,10.0])

proc observeParsingDuration*(scraper: string; seconds: float) =
  observe("kite_parsing_duration_seconds", {("scraper", scraper)}, seconds, @[0.01,0.05,0.1,0.5,1.0,5.0])

# App info
proc appInfo*(): string =
  "kite_app_info{version=\"2.0.0\",name=\"Kite\",description=\"Legal case law scraper library\"} 1\n"

# Exposition
proc metricsText*(): string =
  var s = "# HELP kite_app Application information\n# TYPE kite_app info\n"
  s.add appInfo()

  let counterHelp: Table[string,string] = {
    "kite_requests_total": "Total number of HTTP requests made",
    "kite_request_errors_total": "Total number of request errors",
    "kite_cases_scraped_total": "Total number of cases successfully scraped",
    "kite_rate_limit_hits_total": "Number of times rate limiting was triggered",
    "kite_parsing_errors_total": "Total number of parsing errors",
  }.toTable

  for name, store in counters:
    s.add "# HELP " & name & " " & counterHelp.getOrDefault(name, "") & "\n# TYPE " & name & " counter\n"
    for k, v in store:
      # convert LabelKey canon to {k="v",...}
      var labelsOut: seq[string] = @[]
      for part in k.split(';'):
        if part.len == 0: continue
        let kv = part.split('=')
        if kv.len == 2:
          labelsOut.add kv[0] & "=\"" & kv[1] & "\""
      let labelStr = if labelsOut.len > 0: "{" & labelsOut.join(",") & "}" else: ""
      s.add name & labelStr & " " & $v & "\n"

  # Gauges
  for name, store in gauges:
    s.add "# HELP " & name & " " & (if name == "kite_active_scrapers": "Number of currently active scrapers" else: "") & "\n# TYPE " & name & " gauge\n"
    for k, v in store:
      var labelsOut: seq[string] = @[]
      for part in k.split(';'):
        if part.len == 0: continue
        let kv = part.split('=')
        if kv.len == 2:
          labelsOut.add kv[0] & "=\"" & kv[1] & "\""
      let labelStr = if labelsOut.len > 0: "{" & labelsOut.join(",") & "}" else: ""
      s.add name & labelStr & " " & $v & "\n"

  # Histograms
  let histHelp: Table[string,string] = {
    "kite_request_duration_seconds": "HTTP request duration in seconds",
    "kite_scraping_duration_seconds": "Time taken to scrape cases",
    "kite_rate_limit_wait_seconds": "Time spent waiting due to rate limiting",
    "kite_parsing_duration_seconds": "Time taken to parse case data",
  }.toTable

  for name, h in histos:
    s.add "# HELP " & name & " " & histHelp.getOrDefault(name, "") & "\n# TYPE " & name & " histogram\n"
    for k, counts in h.counts:
      # base labels without le
      var baseLabels: seq[(string,string)] = @[]
      for part in k.split(';'):
        if part.len == 0: continue
        let kv = part.split('=')
        if kv.len == 2: baseLabels.add (kv[0], kv[1])
      # cumulative
      var cum = 0.0
      for i in 0 ..< h.buckets.len:
        cum += counts[i]
        var labelsOut = baseLabels
        labelsOut.add ("le", $h.buckets[i])
        let keyStr = "{" & labelsOut.mapIt(it[0] & "=\"" & it[1] & "\"").join(",") & "}"
        s.add name & "_bucket" & keyStr & " " & $cum & "\n"
      # +Inf
      cum += counts[^1]
      var labelsOut2 = baseLabels
      labelsOut2.add ("le", "+Inf")
      let keyStr2 = "{" & labelsOut2.mapIt(it[0] & "=\"" & it[1] & "\"").join(",") & "}"
      s.add name & "_bucket" & keyStr2 & " " & $cum & "\n"
      # _sum and _count
      let sumv = h.sums.getOrDefault(k, 0.0)
      let cntv = h.totals.getOrDefault(k, 0.0)
      let keyBase = if baseLabels.len > 0: "{" & baseLabels.mapIt(it[0] & "=\"" & it[1] & "\"").join(",") & "}" else: ""
      s.add name & "_sum" & keyBase & " " & $sumv & "\n"
      s.add name & "_count" & keyBase & " " & $cntv & "\n"

  return s

# Simple HTTP server exposing /metrics and /health
proc handleClient(sock: AsyncSocket; metricsPath: string) {.async.} =
  let req = await sock.recvLine()
  if req.startsWith("GET "):
    if req.contains(" " & metricsPath & " "):
      let body = metricsText()
      let resp = "HTTP/1.1 200 OK\r\nContent-Type: text/plain; version=0.0.4\r\nContent-Length: " & $body.len & "\r\n\r\n" & body
      await sock.send(resp)
    elif req.contains(" /health "):
      let body = "{\"status\": \"healthy\", \"service\": \"kite\"}"
      let resp = "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: " & $body.len & "\r\n\r\n" & body
      await sock.send(resp)
    else:
      await sock.send("HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n")
  await sock.close()

proc runMetricsServer*() {.async.} =
  let cfg = getConfig()
  let port = cfg.metricsPort
  let path = cfg.metricsPath
  let server = newAsyncSocket()
  server.setSockOpt(OptReuseAddr, true)
  await server.bindAddr(Port(port))
  server.listen()
  info "metrics_server_started", port=port, metrics_path=path, health_path="/health"
  while true:
    let client = await server.accept()
    asyncCheck handleClient(client, path)
