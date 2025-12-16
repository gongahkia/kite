import std/[osproc, strutils]
import unittest

suite "CLI basic tests":
  test "list-scrapers prints known scrapers":
    let p = startProcess("./kite", args=["list-scrapers"], options={poUsePath})
    let output = p.outputStream.readAll()
    discard p.waitForExit()
    check output.toLowerAscii.contains("courtlistener")
    check output.toLowerAscii.contains("canlii")

  test "search unknown scraper exits":
    let p = startProcess("./kite", args=["search", "unknown", "query"], options={poUsePath, poStdErrToStdOut})
    let rc = p.waitForExit()
    check rc != 0
