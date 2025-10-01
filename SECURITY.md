# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of dbt2lookml seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please do NOT

- Open a public GitHub issue for security vulnerabilities
- Disclose the vulnerability publicly before it has been addressed

### Please DO

**Report security vulnerabilities via GitHub Security Advisories:**

1. Go to the [Security tab](https://github.com/magnus-ffcg/go-dbt2lookml/security) of this repository
2. Click "Report a vulnerability"
3. Fill out the form with details about the vulnerability

**Or email directly:**

- **Email:** [security contact to be added]
- **Response time:** We aim to respond within 48 hours

### What to Include

When reporting a vulnerability, please include:

1. **Description** of the vulnerability
2. **Steps to reproduce** the issue
3. **Potential impact** of the vulnerability
4. **Suggested fix** (if you have one)
5. **Your contact information** for follow-up

### What to Expect

1. **Acknowledgment:** We'll acknowledge receipt within 48 hours
2. **Assessment:** We'll assess the vulnerability and determine its impact
3. **Fix:** We'll develop a fix and test it thoroughly
4. **Disclosure:** We'll coordinate disclosure with you
5. **Release:** We'll release a patched version
6. **Credit:** We'll credit you in the security advisory (unless you prefer to remain anonymous)

## Security Measures

This project implements several security measures:

- **Dependency Scanning:** Automated `govulncheck` runs on every PR and weekly
- **Static Analysis:** `staticcheck` and `golangci-lint` catch potential issues
- **Code Review:** All changes reviewed before merging
- **Minimal Dependencies:** We keep dependencies minimal and well-maintained
- **Input Validation:** All user inputs are validated and sanitized

## Security Best Practices for Users

When using dbt2lookml:

1. **Keep Updated:** Always use the latest version to get security patches
2. **Validate Inputs:** Ensure your dbt manifest and catalog files are from trusted sources
3. **Check Permissions:** Review file permissions for generated LookML files
4. **Audit Dependencies:** Run `go mod verify` to check dependency integrity
5. **Report Issues:** If you find something suspicious, report it!

## Vulnerability Disclosure Policy

- We follow a **responsible disclosure** process
- **Embargo period:** Typically 90 days or until a fix is released
- **Public disclosure:** After fix is released, we publish security advisory
- **CVE assignment:** For significant vulnerabilities, we request CVE IDs
- **Credit:** We publicly credit security researchers (with permission)

## Security Updates

Security updates are published:

1. As new releases on GitHub
2. In the [CHANGELOG.md](CHANGELOG.md)
3. As GitHub Security Advisories
4. Via GitHub watch notifications (if you're watching the repo)

## Contact

For security-related questions or concerns:

- **GitHub Security:** Use GitHub Security Advisories (preferred)
- **Email:** [To be added - security contact]
- **Response Time:** Within 48 hours for security issues

## Acknowledgments

We thank all security researchers who responsibly disclose vulnerabilities. Your contributions help keep dbt2lookml and its users safe.

---

**Last Updated:** 2025-10-01
