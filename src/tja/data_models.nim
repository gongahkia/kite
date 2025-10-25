import std/[times, json]

type
  CaseData* = object
    caseName*: string
    caseId*: string
    court*: string
    date*: DateTime
    url*: string
    fullText*: string
    summary*: string
    judges*: seq[string]
    parties*: seq[string]
    legalIssues*: seq[string]
    citations*: seq[string]
    jurisdiction*: string
    caseType*: string
    outcome*: string
    metadata*: JsonNode

proc initCaseData*(): CaseData =
  result.caseName = ""
  result.caseId = ""
  result.court = ""
  result.url = ""
  result.fullText = ""
  result.summary = ""
  result.judges = @[]
  result.parties = @[]
  result.legalIssues = @[]
  result.citations = @[]
  result.jurisdiction = ""
  result.caseType = ""
  result.outcome = ""
  result.metadata = newJObject()

proc toJson*(c: CaseData): JsonNode =
  let dateStr = if c.date.year != 0: %$c.date else: newJNull()
  result = newJObject()
  result["case_name"] = % c.caseName
  result["case_id"] = % c.caseId
  result["court"] = % c.court
  result["date"] = dateStr
  result["url"] = % c.url
  result["full_text"] = % c.fullText
  result["summary"] = % c.summary
  result["judges"] = % c.judges
  result["parties"] = % c.parties
  result["legal_issues"] = % c.legalIssues
  result["citations"] = % c.citations
  result["jurisdiction"] = % c.jurisdiction
  result["case_type"] = % c.caseType
  result["outcome"] = % c.outcome
  result["metadata"] = if c.metadata.isNil: newJObject() else: c.metadata
