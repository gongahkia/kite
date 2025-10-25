import std/[json, strutils, times]
import chronicles
from ./data_models import CaseData, toJson
from ./scrapers/courtlistener import CourtListenerScraper, newCourtListenerScraper, searchCases as clSearch, getCaseById as clGet
from ./scrapers/canlii import CanLIIScraper, newCanLIIScraper, searchCases as caSearch, getCaseById as caGet
from ./scrapers/austlii import AustLIIScraper, newAustLIIScraper, searchCases as auSearch, getCaseById as auGet
from ./scrapers/bailii import BAILIIScraper, newBAILIIScraper, searchCases as baSearch, getCaseById as baGet
from ./scrapers/findlaw import FindLawScraper, newFindLawScraper, searchCases as flSearch, getCaseById as flGet
from ./scrapers/singapore_judiciary import SingaporeJudiciaryScraper, newSingaporeJudiciaryScraper, searchCases as sgSearch, getCaseById as sgGet
from ./scrapers/indian_kanoon import IndianKanoonScraper, newIndianKanoonScraper, searchCases as ikSearch, getCaseById as ikGet
from ./scrapers/hklii import HKLIIScraper, newHKLIIScraper, searchCases as hkSearch, getCaseById as hkGet
from ./scrapers/legifrance import LegifranceScraper, newLegifranceScraper, searchCases as lfSearch, getCaseById as lfGet
from ./scrapers/german_law_archive import GermanLawArchiveScraper, newGermanLawArchiveScraper, searchCases as deSearch, getCaseById as deGet
from ./scrapers/curia_europa import CuriaEuropaScraper, newCuriaEuropaScraper, searchCases as euSearch, getCaseById as euGet
from ./scrapers/worldlii import WorldLIIScraper, newWorldLIIScraper, searchCases as wlSearch, getCaseById as wlGet
from ./scrapers/worldcourts import WorldCourtsScraper, newWorldCourtsScraper, searchCases as wcSearch, getCaseById as wcGet
from ./scrapers/supremecourt_india import SupremeCourtIndiaScraper, newSupremeCourtIndiaScraper, searchCases as siSearch, getCaseById as siGet
from ./scrapers/kenya_law import KenyaLawScraper, newKenyaLawScraper, searchCases as keSearch, getCaseById as keGet
from ./scrapers/supremecourt_japan import SupremeCourtJapanScraper, newSupremeCourtJapanScraper, searchCases as jpSearch, getCaseById as jpGet
from ./scrapers/legal_tools import LegalToolsScraper, newLegalToolsScraper, searchCases as ltSearch, getCaseById as ltGet

proc formatCase*(c: CaseData; format = "text"): string =
  case format
  of "json": result = $toJson(c)
  of "csv":
    result = "\"" & c.caseName & "\",\"" & c.caseId & "\",\"" & c.court & "\",\"" & $c.date & "\",\"" & c.url & "\",\"" & c.jurisdiction & "\""
  else:
    result = "Case: " & c.caseName & "\n"
    result.add "ID: " & c.caseId & "\n"
    result.add "Court: " & c.court & "\n"
    result.add "Date: " & $c.date & "\n"
    result.add "URL: " & c.url & "\n"
    result.add "Jurisdiction: " & c.jurisdiction & "\n"
    result.add "--------------------------------------------------------------------------------"

proc listScrapersInfo*(): seq[(string,string)] = @[
  ("courtlistener", "CourtListener (US federal and state case law)"),
  ("findlaw", "FindLaw (US Supreme Court and state case law)"),
  ("austlii", "AustLII (Australian case law)"),
  ("canlii", "CanLII (Canadian case law)"),
  ("bailii", "BAILII (UK and Ireland case law)"),
  ("singapore", "Singapore Judiciary (Singapore case law)"),
  ("indian-kanoon", "Indian Kanoon (Indian case law)"),
  ("hklii", "HKLII (Hong Kong case law)"),
  ("legifrance", "Légifrance (French case law)"),
  ("german-law", "German Law Archive (German case law)"),
  ("curia-europa", "Curia Europa (EU case law)"),
  ("worldlii", "WorldLII (International case law)"),
  ("worldcourts", "WorldCourts (International case law)"),
  ("supremecourt-india", "Supreme Court of India"),
  ("kenya-law", "Kenya Law (Kenyan case law)"),
  ("supremecourt-japan", "Supreme Court of Japan"),
  ("legal-tools", "ICC Legal Tools Database")
]

proc runSearch*(scraper, query: string; startDate = ""; endDate = ""; court = ""; limit = 20; format = "text"; output = "") =
  var cases: seq[CaseData] = @[]
  case scraper
  of "courtlistener": cases = newCourtListenerScraper().clSearch(query, startDate, endDate, court, limit)
  of "canlii": cases = newCanLIIScraper().caSearch(query, startDate, endDate, court, limit)
  of "austlii": cases = newAustLIIScraper().auSearch(query, startDate, endDate, court, limit)
  of "bailii": cases = newBAILIIScraper().baSearch(query, startDate, endDate, court, limit)
  of "findlaw": cases = newFindLawScraper().flSearch(query, startDate, endDate, court, limit)
  of "singapore": cases = newSingaporeJudiciaryScraper().sgSearch(query, startDate, endDate, court, limit)
  of "indian-kanoon": cases = newIndianKanoonScraper().ikSearch(query, startDate, endDate, court, limit)
  of "hklii": cases = newHKLIIScraper().hkSearch(query, startDate, endDate, court, limit)
  of "legifrance": cases = newLegifranceScraper().lfSearch(query, startDate, endDate, court, limit)
  of "german-law": cases = newGermanLawArchiveScraper().deSearch(query, startDate, endDate, court, limit)
  of "curia-europa": cases = newCuriaEuropaScraper().euSearch(query, startDate, endDate, court, limit)
  of "worldlii": cases = newWorldLIIScraper().wlSearch(query, startDate, endDate, court, limit)
  of "worldcourts": cases = newWorldCourtsScraper().wcSearch(query, startDate, endDate, court, limit)
  of "supremecourt-india": cases = newSupremeCourtIndiaScraper().siSearch(query, startDate, endDate, court, limit)
  of "kenya-law": cases = newKenyaLawScraper().keSearch(query, startDate, endDate, court, limit)
  of "supremecourt-japan": cases = newSupremeCourtJapanScraper().jpSearch(query, startDate, endDate, court, limit)
  of "legal-tools": cases = newLegalToolsScraper().ltSearch(query, startDate, endDate, court, limit)
  else: quit("Error: Unknown scraper '" & scraper & "'", 1)

  if cases.len == 0:
    echo "No cases found."
    return
  echo "Found " & $cases.len & " cases:\n"
  for c in cases:
    let out = formatCase(c, format)
    echo out

proc runGetCase*(scraper, caseId: string; format = "text"; output = "") =
  var c: CaseData
  case scraper
  of "courtlistener": c = newCourtListenerScraper().clGet(caseId)
  of "canlii": c = newCanLIIScraper().caGet(caseId)
  of "austlii": c = newAustLIIScraper().auGet(caseId)
  of "bailii": c = newBAILIIScraper().baGet(caseId)
  of "findlaw": c = newFindLawScraper().flGet(caseId)
  of "singapore": c = newSingaporeJudiciaryScraper().sgGet(caseId)
  of "indian-kanoon": c = newIndianKanoonScraper().ikGet(caseId)
  of "hklii": c = newHKLIIScraper().hkGet(caseId)
  of "legifrance": c = newLegifranceScraper().lfGet(caseId)
  of "german-law": c = newGermanLawArchiveScraper().deGet(caseId)
  of "curia-europa": c = newCuriaEuropaScraper().euGet(caseId)
  of "worldlii": c = newWorldLIIScraper().wlGet(caseId)
  of "worldcourts": c = newWorldCourtsScraper().wcGet(caseId)
  of "supremecourt-india": c = newSupremeCourtIndiaScraper().siGet(caseId)
  of "kenya-law": c = newKenyaLawScraper().keGet(caseId)
  of "supremecourt-japan": c = newSupremeCourtJapanScraper().jpGet(caseId)
  of "legal-tools": c = newLegalToolsScraper().ltGet(caseId)
  else: quit("Error: Unknown scraper '" & scraper & "'", 1)

  if c.caseName.len == 0:
    echo "Case '" & caseId & "' not found."
    return
  echo formatCase(c, format)
