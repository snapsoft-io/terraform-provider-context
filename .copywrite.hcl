schema_version = 1

project {
  license          = "MIT"
  copyright_holder = "SnapSoft"
  copyright_year   = 2026

  header_ignore = [
    # examples used within documentation (prose)
    "examples/**",

    # golangci-lint tooling configuration
    ".golangci.yml",

    # GoReleaser tooling configuration
    ".goreleaser.yml",
  ]
}
