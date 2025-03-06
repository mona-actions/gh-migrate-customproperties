# gh-migrate-customproperties

A GitHub CLI extension that helps migrate repository custom properties from one organization to another. It supports both GitHub.com and GitHub Enterprise Server environments.


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
  -t, --target-organization string   Target Organization to sync properties to
  -a, --source-token string         Source Organization GitHub token. Required scopes: read:org, read:user, user:email
  -b, --target-token string         Target Organization GitHub token. Required scopes: admin:org
  -u, --source-hostname string      GitHub Enterprise source hostname url (optional) Ex. https://github.example.com
  -r, --repository-list string      File containing list of repositories to sync properties from. One repository per line.
```

### Repository List Format

The repository list file (`--repository-list`) must contain repositories in either of these formats:
- Full URLs: `https://github.com/owner/repo`
- Owner/repo format: `owner/repo`

Example repository list:
```
octocat/Hello-World
https://github.com/mona/awesome-project
```

Note: Simple repository names without owner are no longer supported. Each entry must specify both the owner and repository name.

## License

- [MIT](./LICENSE) (c) [Mona-Actions](https://github.com/mona-actions)
- [Contributing](./CONTRIBUTING.md)
