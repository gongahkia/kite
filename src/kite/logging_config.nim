import chronicles
import std/[strutils]
from ./config import getConfig

var configured = false

proc configureLogger*() =
  if configured: return
  let cfg = getConfig()
  when declared(setLogLevel):
    case cfg.logLevel.toUpperAscii
    of "DEBUG": setLogLevel(LogLevel.debug)
    of "INFO": setLogLevel(LogLevel.info)
    of "WARNING": setLogLevel(LogLevel.warn)
    of "ERROR": setLogLevel(LogLevel.error)
    of "CRITICAL": setLogLevel(LogLevel.fatal)
    else: setLogLevel(LogLevel.info)
  # chronicles prints JSON-like with structured fields by default; ensure key=value pairs
  configured = true

proc getLogger*(_: typedesc): auto =
  configureLogger()
  return defaultLogger()
