# gh-atat

CLI tool for synchronizing TODO.md with GitHub Issues

## Motivation

Managing tasks between local TODO.md files and GitHub Issues often leads to duplication and synchronization problems. gh-atat solves this by providing a git-like workflow to keep both in sync automatically.

## Prerequisites

- GitHub CLI (`gh`) installed and authenticated

## Installation

```bash
gh extension install toms74209200/gh-atat
```

## Usage

### Repository Setup

Add a repository to sync with:

```bash
gh atat remote add owner/repo
```

View current repository configuration:

```bash
gh atat remote
```

Remove a repository:

```bash
gh atat remote remove owner/repo
```

### Commands

Push TODO.md to GitHub Issues

```bash
gh atat push
```

Pull GitHub Issues to TODO.md

```bash
gh atat pull
```

### TODO.md Format

gh-atat works with standard markdown checkbox format:

```markdown
- [ ] Implement new feature
- [x] Fix bug in authentication
- [ ] Update documentation
```

After synchronization, Issue numbers will be automatically added:

```markdown
- [ ] Implement new feature #123
- [x] Fix bug in authentication #124
- [ ] Update documentation #125
```

## License

[MIT License](LICENSE)

## Author

[toms74209200](https://github.com/toms74209200)
