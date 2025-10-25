import std/[os, strutils]

type
  Config* = object
    pythonEnv*: string
    logLevel*: string
    rateLimit*: float
    requestTimeout*: int
    maxRetries*: int
    userAgent*: string
    metricsPort*: int
    metricsPath*: string
    enableMetrics*: bool
    enableTracing*: bool
    logFormat*: string
    logFilePath*: string
    logRotation*: string
    logRetention*: int

var _config*: Config
var _initialized = false

proc parseBoolEnv(key: string; defaultVal: bool): bool =
  let v = getEnv(key, if defaultVal: "true" else: "false").toLowerAscii
  v == "true" or v == "1" or v == "yes"

proc getConfig*(): Config =
  if not _initialized:
    _config.pythonEnv = getEnv("PYTHON_ENV", "development")
    _config.logLevel = getEnv("LOG_LEVEL", "INFO")
    _config.rateLimit = parseFloat(getEnv("RATE_LIMIT", "1.0"))
    _config.requestTimeout = parseInt(getEnv("REQUEST_TIMEOUT", "30"))
    _config.maxRetries = parseInt(getEnv("MAX_RETRIES", "3"))
    _config.userAgent = getEnv("USER_AGENT", "")
    _config.metricsPort = parseInt(getEnv("METRICS_PORT", "8000"))
    _config.metricsPath = getEnv("METRICS_PATH", "/metrics")
    _config.enableMetrics = parseBoolEnv("ENABLE_METRICS", true)
    _config.enableTracing = parseBoolEnv("ENABLE_TRACING", true)
    _config.logFormat = getEnv("LOG_FORMAT", "json")
    _config.logFilePath = getEnv("LOG_FILE_PATH", "./logs/app.log")
    _config.logRotation = getEnv("LOG_ROTATION", "10MB")
    _config.logRetention = parseInt(getEnv("LOG_RETENTION", "30"))
    # basic validation
    if _config.requestTimeout <= 0: _config.requestTimeout = 30
    if _config.maxRetries < 0: _config.maxRetries = 0
    if _config.metricsPort < 1 or _config.metricsPort > 65535: _config.metricsPort = 8000
    _initialized = true
  return _config

proc isProduction*(c: Config): bool = c.pythonEnv.toLowerAscii == "production"
proc isDevelopment*(c: Config): bool = c.pythonEnv.toLowerAscii == "development"

proc resetConfig*() =
  _initialized = false
