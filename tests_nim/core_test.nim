import std/[unittest, json, times]
from ../src/tja/helpers import sanitizeText, normalizeCourtName
from ../src/tja/data_models import CaseData, initCaseData, toJson

suite "Core helpers and models":
  test "sanitizeText collapses whitespace":
    check sanitizeText("a   b\n\t c") == "a b c"

  test "normalizeCourtName expands abbreviations":
    check normalizeCourtName("S.C.").contains("Supreme Court")

  test "CaseData toJson has expected keys":
    var c = initCaseData()
    c.caseName = "Foo v. Bar"
    let j = toJson(c)
    check j.hasKey("case_name")
    check j["case_name"].getStr == "Foo v. Bar"
