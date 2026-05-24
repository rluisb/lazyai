# Pull Requests

```
blog:  -- changes to our documentation found in the `docs/blog` directory
docs:  -- changes to our documentation found in the `docs/` directory, except blogs
feat:  -- changes that add a new feature
fix:   -- changes that fix existing functionality
refactor: -- changes that refactor existing functionality without adding any new features
feat!:/refactor!:/fix!: -- same as above, but indicates a breaking change. Examples would be changes to the public C API, renaming a board/shield, editing a board or shield to rename devicetree labels that may be used in keymaps, etc.
ci:   -- changes to our continuous integration setup with GitHub Actions, usually only for the files in `.github/workflows/`
chore: -- grab bag type for small changes that don't fall into any of the above categories, including dependency updates for development tools and docs.
```

## Opening a PR

Create a PR by visiting and clicking "New pull request" and selecting your branch. GitHub should auto-populate a start of a description/body to your PR, but please make sure to supplement that with additional details about the change and make sure the check-list is complete in the PR template.

Once created, the PR should automatically have reviewers assigned.

## Review

Depending on the area of change, different ZMK team members will review the PR. The ZMK project is a small team, so please be patient with the review timeline. You are welcome to send a polite request/nudge on our Discord server to draw attention to your PR if it has not gotten a response after a reasonable amount of time.

## Merging

Maintainers merging PRs will perform the following steps: