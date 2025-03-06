# gh-migrate-customproperties

A GitHub CLI extension that helps migrate repository custom properties from one organization to another. It supports both GitHub.com and GitHub Enterprise Server environments for source.

## Installation

```bash
gh extension install mona-actions/gh-migrate-customproperties
```

## Upgrade

To upgrade the extension to the latest version:

```bash
gh extension upgrade gh-migrate-customproperties
```

## Usage

```bash
gh migrate-customproperties [flags]

Flags:
  -s, --source-organization string   Source Organization to sync custom properties from
  -t, --target-organization string   Target Organization to sync custom properties to
  -a, --source-token string         Source Organization GitHub token. Required scopes: read:org, read:user, user:email
  -b, --target-token string         Target Organization GitHub token. Required scopes: admin:org
  -u, --source-hostname string      GitHub Enterprise source hostname url (optional) Ex. https://github.example.com
  -r, --repository-list string      File containing list of repositories to sync properties from. One repository per line.
```

### Example repository list file format

The repository list file (`--repository-list`) should contain one repository per line and supports the following formats:
- Full URLs: `https://github.com/org/repo1` (the `source-organization` and `source-hostname` value will be used for the org and hostname respectively)
- Simple org/repo format: `org/repo1` (the `source-organization` value will be used for the org)
- Just repository names: `repo1`

## License

- [MIT](./LICENSE) (c) [Mona-Actions](https://github.com/mona-actions)
- [Contributing](./CONTRIBUTING.md)
