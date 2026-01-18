# Security Policy

## Supported Versions

Security fixes are applied to the latest release on the `main` branch.

## Reporting a Vulnerability

Please **do not** report security vulnerabilities through public GitHub issues, discussions, or pull requests.

Instead, report them privately:

1. Go to this repository’s **Security** tab and use **“Report a vulnerability”** (preferred).
2. If that is not available for any reason, contact the maintainer directly via the contact method listed on the maintainer’s GitHub profile.

When reporting, please include:
- A clear description of the vulnerability and impact
- Steps to reproduce (proof-of-concept if possible)
- Any affected endpoint(s) or configuration
- Your assessment of severity

### What counts as a security issue in this project?

Because this project handles secrets, security issues include (but are not limited to):
- Secrets retrievable after they should have been "burned" / deleted
- Disclosure through logs, error messages, templates, or static assets
- Authentication/authorization bypass (if/when auth features exist)
- Weak randomness / predictable secret identifiers
- Insecure storage or unsafe defaults
- SSRF, XSS, CSRF, request smuggling, header injection

## Disclosure Process

After receiving a report, maintainers will:
1. Confirm receipt and begin triage.
2. Evaluate impact and scope.
3. Work on a fix and regression tests.
4. Release a patched version.

We may request additional information to reproduce and validate the issue.

---

Tracking issue: #126
