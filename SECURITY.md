# Security Policy

## Supported Versions

Only the latest major version (v1.x.x) is supported with security updates.

| Version | Supported |
| ------- | --------- |
| 1.x.x   | ✅        |
| < 1.0   | ❌        |

## Scope

`acell` is a low-level terminal library: it switches the terminal to raw mode,
parses keyboard/mouse ANSI sequences, and writes ANSI escape sequences to the
output. The following are considered in scope:

- Memory safety issues in ANSI/input parsers (`ansi`, `term`).
- Out-of-bounds writes in `Flush` or buffer helpers (`NewBuf`).
- Cases where a malformed input sequence causes a panic, hang, or unbounded
  memory growth.
- Terminal escape sequences that could be used for injection through a
  widget's `Cell.Char` or `Style` (e.g. if user-controlled text ends up
  unescaped in the output stream).

The following are **out of scope**:

- Behavior of the host terminal emulator itself.
- Denial of service through a user's own infinite render loop.
- Issues in third-party terminal emulators, SSH clients, or shells.

## Reporting a Vulnerability

Please report security vulnerabilities by opening a **private vulnerability report** on GitHub:

1. Go to the **Security** tab of this repository.
2. Click **Report a vulnerability**.
3. Fill in the details.

Alternatively, you can send an email to **r@createlec.com**.

I will respond within **7 days** and coordinate a fix and disclosure timeline.