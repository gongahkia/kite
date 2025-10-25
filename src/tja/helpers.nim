import std/[strutils, re, times, uri, unicode]

proc validateDate*(dateInput: string): DateTime =
  let s = dateInput.strip
  if s.len == 0: return default(DateTime)
  if s.contains('-') and s.len >= 10:
    return parse("yyyy-MM-dd", s)
  if s.contains('/') and s.len >= 8:
    try:
      return parse("dd/MM/yyyy", s)
    except TimeParseError:
      discard
  return parseISO(s)

proc sanitizeText*(text: string): string =
  if text.len == 0: return ""
  var t = text
  t = t.replace(re"\s+", " ")
  t = t.multiReplace({"&nbsp;": " ", "&amp;": "&", "&lt;": "<", "&gt;": ">", "&quot;": "\"", "&#39;": "'", "&apos;": "'"})
  t = t.replace("\f", " ")
  t = t.replace("\x0c", " ")
  t = t.replace("\u00a0", " ")
  t = t.replace(re"\s+", " ").strip
  return t

proc normalizeCourtName*(court: string): string =
  if court.len == 0: return ""
  var r = court.strip
  r = r.replace(re"\bS\.?C\.?\b", "Supreme Court")
  r = r.replace(re"\bC\.?A\.?\b", "Court of Appeal")
  r = r.replace(re"\bH\.?C\.?\b", "High Court")
  r = r.replace(re"\bD\.?C\.?\b", "District Court")
  r = r.replace(re"\bF\.?C\.?\b", "Federal Court")
  r = r.replace(re"\bCt\.?\b", "Court")
  r = r.replace(re"\bJ\.?\b", "Justice")
  return r.strip

proc buildSearchUrl*(baseUrl: string; params: Table[string, string]): string =
  if params.len == 0: return baseUrl
  var q: seq[(string, string)] = @[]
  for k, v in params.pairs:
    if v.len > 0: q.add (k, v)
  if q.len == 0: return baseUrl
  let sep = if baseUrl.contains('?'): '&' else: '?'
  result = baseUrl & $sep & encodeQuery(q)

# Unified date normalization across locales
proc normalizeDate*(input: string): DateTime =
  let s = sanitizeText(input)
  if s.len == 0: return default(DateTime)
  # Try ISO first
  try: return parse("yyyy-MM-dd", s) except: discard
  # Try common formats
  for fmt in ["dd/MM/yyyy", "d/MM/yyyy", "dd/MM/yy", "d MMMM yyyy", "dd MMM yyyy", "MMMM d, yyyy", "dd.MM.yyyy", "d.M.yyyy"]:
    try: return parse(fmt, s) except: discard
  # English months already handled by formats; handle French months
  var fr = s.toLowerAscii
  let frMap = {
    "janvier": "January", "février": "February", "fevrier": "February", "mars": "March", "avril": "April", "mai": "May", "juin": "June", "juillet": "July", "août": "August", "aout": "August", "septembre": "September", "octobre": "October", "novembre": "November", "décembre": "December", "decembre": "December"
  }
  for k, v in frMap:
    if k in fr:
      fr = fr.replace(k, v.toLowerAscii)
  for fmt in ["d mmmm yyyy", "dd mmmm yyyy"]:
    try: return parse(fmt, fr) except: discard
  # German months
  var de = s
  let deMap = {
    "Januar": "January", "Februar": "February", "März": "March", "Maerz": "March", "April": "April", "Mai": "May", "Juni": "June", "Juli": "July", "August": "August", "September": "September", "Oktober": "October", "November": "November", "Dezember": "December"
  }
  for k, v in deMap:
    if k in de:
      de = de.replace(k, v)
  for fmt in ["d MMMM yyyy", "dd MMMM yyyy", "dd.MM.yyyy", "d.M.yyyy"]:
    try: return parse(fmt, de) except: discard
  # Japanese yyyy年mm月dd日
  let jpRe = re"(\d{4})年(\d{1,2})月(\d{1,2})日"
  let m = jpRe.find(s)
  if m.matchedLen > 0:
    let y = parseInt(s.substr(m.group(0,1).a, m.group(0,1).b))
    let mo = parseInt(s.substr(m.group(0,2).a, m.group(0,2).b))
    let d = parseInt(s.substr(m.group(0,3).a, m.group(0,3).b))
    return initDateTime(d, Month(mo), y, 0,0,0, local())
  # Fallback: try ISO parser
  return parseISO(s)
