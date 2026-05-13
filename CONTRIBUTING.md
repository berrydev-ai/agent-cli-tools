# Contributing

Thanks for contributing to Agent CLI Tools.

## Development

1. Add each CLI under `cmd/<tool>`.
2. Keep tool behavior in focused packages under `internal/cli/<tool>`.
3. Support flags and environment variables for required inputs.
4. Add or update command tests for behavior changes.
5. Update the tool README and `.env.example` when inputs, output, or usage changes.

Run tests before opening a pull request:

```bash
make test
make build
```

## Releases

Each binary is released independently with a per-binary SemVer tag:

```text
<tool>-v<major>.<minor>.<patch>
```

Use the Makefile targets so validation, archive names, and tags stay consistent:

```bash
make print-tag TOOL=slack-post VERSION=v0.1.0
make release-archives TOOL=slack-post VERSION=v0.1.0
make tag TOOL=slack-post VERSION=v0.1.0
make push-tag TOOL=slack-post VERSION=v0.1.0
```

Pushing the tag runs the release workflow, publishes platform archives, and generates release notes from merged pull requests and commits. Release note categories are configured in `.github/release.yml`.

## License

By contributing, you agree that your contributions are licensed under the MIT License.
