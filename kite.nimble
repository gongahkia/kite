version       = "2.0.0"
author        = "Kite Contributors"
description   = "Nim rewrite of Kite - legal case law scrapers"
license       = "MIT"
srcDir        = "src"
bin           = @[("kite", "src/kite.nim")]

requires "nim >= 2.0.0"
requires "cligen >= 1.6.0"
requires "chronicles >= 0.11.5"
requires "httpx >= 0.7.5"
requires "jsony >= 1.1.5"
requires "regex >= 0.20.3"

task test, "Run tests":
  exec "nim c -r tests_nim/cli_test.nim"

task docs, "Generate documentation":
  exec "nim doc --project --outdir:htmldocs src/kite.nim"
