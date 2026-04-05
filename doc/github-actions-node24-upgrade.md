# Fixing GitHub Actions Node.js 20 Deprecation Warnings

## The Problem

GitHub Actions is deprecating Node.js 20. Starting June 2, 2026, actions will be forced to run on Node.js 24. Starting September 16, 2026, Node.js 20 will be removed entirely.

You'll see a warning like:

> Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, actions/setup-go@v5, actions/upload-artifact@v4.

## The Fix

Update your workflow YAML files (`.github/workflows/*.yml`) to use the latest major versions of each action. These versions run on Node.js 24.

### Action Version Mapping

| Action | Old (Node 20) | New (Node 24) |
|---|---|---|
| `actions/checkout` | v4 | **v6** |
| `actions/setup-go` | v5 | **v6** |
| `actions/setup-node` | v4 | **v5** |
| `actions/setup-python` | v5 | **v6** |
| `actions/setup-java` | v4 | **v5** |
| `actions/upload-artifact` | v4 | **v7** |
| `actions/download-artifact` | v4 | **v8** |
| `actions/cache` | v4 | **v5** |
| `actions/github-script` | v7 | **v8** |

Versions current as of April 2026. Check the releases page for each action if in doubt: `https://github.com/{action}/releases`

### Example

Before:
```yaml
steps:
  - uses: actions/checkout@v4
  - uses: actions/setup-go@v5
  - uses: actions/upload-artifact@v4
```

After:
```yaml
steps:
  - uses: actions/checkout@v6
  - uses: actions/setup-go@v6
  - uses: actions/upload-artifact@v7
```

### Quick sed commands

To update all workflow files at once:

```sh
cd .github/workflows
sed -i 's|actions/checkout@v4|actions/checkout@v6|g' *.yml
sed -i 's|actions/setup-go@v5|actions/setup-go@v6|g' *.yml
sed -i 's|actions/setup-node@v4|actions/setup-node@v5|g' *.yml
sed -i 's|actions/setup-python@v5|actions/setup-python@v6|g' *.yml
sed -i 's|actions/setup-java@v4|actions/setup-java@v5|g' *.yml
sed -i 's|actions/upload-artifact@v4|actions/upload-artifact@v7|g' *.yml
sed -i 's|actions/download-artifact@v4|actions/download-artifact@v8|g' *.yml
sed -i 's|actions/cache@v4|actions/cache@v5|g' *.yml
sed -i 's|actions/github-script@v7|actions/github-script@v8|g' *.yml
```

### Breaking Changes to Watch For

- **upload-artifact v5+**: The `retention-days` default changed. If you relied on the old default, set it explicitly.
- **download-artifact v5+**: The `merge-multiple` option was added in v5. If you're upgrading from v4, this is a new feature, not a breaking change.
- **checkout v5+**: Sparse checkout options changed. If you use `sparse-checkout`, check the release notes.

Most workflows that use only basic features (checkout, build, upload) will work with no changes beyond the version bump.
