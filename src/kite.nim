import std/[strformat, times, os, json]
import cligen
import chronicles
from tja/cli_adapter import runSearch, runGetCase, listScrapersInfo

proc version*() {.inline.} =
  echo "Kite 2.0.0 (Nim)"

proc cmdListScrapers*() =
  echo "Available scrapers:"
  echo "=".repeat(50)
  for (name, desc) in listScrapersInfo():
    echo name.alignLeft(20), " - ", desc

proc cmdSearch*(scraper, query: string; startDate = ""; endDate = ""; court = ""; limit = 20; format = "text"; output = "") =
  info "search", scraper=scraper, query=query
  runSearch(scraper, query, startDate, endDate, court, limit, format, output)

proc cmdGetCase*(scraper, caseId: string; format = "text"; output = "") =
  info "get_case", scraper=scraper, caseId=caseId
  runGetCase(scraper, caseId, format, output)

when isMainModule:
  dispatchMulti([
    (cmdSearch, "search", short = {"startDate": 'S', "endDate": 'E', "court": 'c', "limit": 'l', "format": 'f', "output": 'o'}),
    (cmdGetCase, "get-case", short = {"format": 'f', "output": 'o'}),
    (cmdListScrapers, "list-scrapers"),
    (version, "--version")
  ],
  help = {
    "search": "Search for cases",
    "get-case": "Get specific case by ID",
    "list-scrapers": "List available scrapers"
  },
  usage = "kite [subcommand] [options]",
  name = "kite")
