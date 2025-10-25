import std/[times, strutils, json]
import httpx
import chronicles
from ./exceptions import ScrapingError, NetworkError, RateLimitError, ParsingError, AuthenticationError
from ./config import getConfig
from ./metrics import incRequest, observeRequestDuration, incRequestError, incRateLimitHits, observeRateLimitWait


type
  BaseScraper* = ref object of RootObj
    rateLimit*: float
    timeout*: int
    maxRetries*: int
    retryDelay*: float
    userAgent*: string
    lastRequestAt: float

method baseUrl*(self: BaseScraper): string {.base.}
method jurisdiction*(self: BaseScraper): string {.base.}

proc defaultUserAgent(): string =
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118 Safari/537.36"

proc initBaseScraper*(rateLimit = 1.0; timeout = 30; maxRetries = 3; retryDelay = 1.0; userAgent = ""): BaseScraper =
  let cfg = getConfig()
  new(result)
  result.rateLimit = rateLimit
  result.timeout = timeout
  result.maxRetries = maxRetries
  result.retryDelay = retryDelay
  result.userAgent = if userAgent.len > 0: userAgent else: (if cfg.userAgent.len > 0: cfg.userAgent else: defaultUserAgent())
  result.lastRequestAt = 0

proc scraperName(self: BaseScraper): string =
  result = self.type.name

proc respectRateLimit(self: BaseScraper) =
  if self.rateLimit > 0:
    let elapsed = epochTime() - self.lastRequestAt
    if elapsed < self.rateLimit:
      let sleepSecs = self.rateLimit - elapsed
      incRateLimitHits(self.scraperName())
      observeRateLimitWait(self.scraperName(), sleepSecs)
      info "rate_limiting", elapsed=elapsed, sleep_time=sleepSecs
      sleep(int(sleepSecs * 1000).Milliseconds)

proc makeRequest*(self: BaseScraper; url: string; method = "GET"; params: seq[(string,string)] = @[]; data = newJObject(); headers: seq[(string,string)] = @[]): Response =
  self.respectRateLimit()
  var reqHeaders = @[ ("User-Agent", self.userAgent), ("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"), ("Accept-Language", "en-US,en;q=0.5"), ("Connection", "keep-alive") ]
  for h in headers: reqHeaders.add h
  var attempts = 0
  while attempts <= self.maxRetries:
    let t0 = epochTime()
    try:
      attempts.inc
      info "making_request", method=method, url=url, attempt=attempts
      var client = newClient(timeout = self.timeout.seconds)
      defer: client.close()
      var resp: Response
      case method
      of "GET": resp = client.get(url, headers=reqHeaders, params=params)
      of "POST": resp = client.post(url, headers=reqHeaders, body=$data)
      else: resp = client.request(method, url, headers=reqHeaders)
      let dt = epochTime() - t0
      observeRequestDuration(self.scraperName(), method, dt)
      self.lastRequestAt = epochTime()
      incRequest(self.scraperName(), method, $resp.code.int)
      if resp.code == Http200:
        return resp
      elif resp.code.int == 429:
        raise newException(RateLimitError, "Rate limited (429)")
      elif resp.code.int in [401, 403]:
        raise newException(AuthenticationError, "Authentication required")
      elif resp.code.int >= 500:
        if attempts <= self.maxRetries:
          let wait = self.retryDelay * (2.0 ** float(attempts-1))
          warn "server_error_retry", status=resp.code.int, retry_delay=wait
          sleep(int(wait * 1000).Milliseconds)
          continue
        else:
          raise newException(NetworkError, "Server error: " & $resp.code.int)
      else:
        raise newException(NetworkError, "HTTP " & $resp.code.int)
    except HttpRequestError as e:
      incRequestError(self.scraperName(), "HttpRequestError")
      if attempts <= self.maxRetries:
        let wait = self.retryDelay * (2.0 ** float(attempts-1))
        warn "connection_error_retry", error=$e.msg, retry_delay=wait
        sleep(int(wait * 1000).Milliseconds)
        continue
      else:
        raise newException(NetworkError, "Connection failed: " & e.msg)
  raise newException(NetworkError, "Failed after retries")

proc parseHtml*(content: string): string = content
