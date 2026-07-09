# Git hooks

Version-controlled git hooks for this repo. They are **not** active until you
point git at this directory (git does not run tracked hooks automatically):

```sh
git config core.hooksPath .githooks
```

Run that once per clone.

## pre-commit — secret scan

Runs [gitleaks](https://github.com/gitleaks/gitleaks) against your **staged**
changes and aborts the commit if it finds a secret (API key, token, private
key, etc.). This catches secrets before they ever enter a commit.

Requirements: `gitleaks` on PATH (`brew install gitleaks`).

- Documented false positives are listed in [`.gitleaksignore`](../.gitleaksignore)
  (by `file:rule:line` fingerprint).
- Emergency bypass: `git commit --no-verify` — avoid it; only when you are
  certain the flagged content is not a real secret and cannot be ignored
  cleanly.
