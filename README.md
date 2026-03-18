# mailgo

`mailgo` is a Go CLI tool that verifies whether an email address is deliverable by running syntax checks and a live SMTP handshake — giving you a clear verdict before you send.

<!-- TODO: add a demo GIF here (e.g. with vhs or asciinema) -->

---

## Motivation

A friend of mine sends cold emails — a lot of them. He kept running into the same problem: a chunk of every list he built was dead weight — bad addresses, catch-all domains that silently swallow messages, and disposable inboxes that go nowhere. He was burning his sender reputation on addresses that were never going to convert.

I looked around for a simple tool he could run from the terminal before importing a list into his outreach tool. Everything I found was either a paid SaaS, a bloated library, or it only checked syntax. So I built `mailgo`: a single binary that does a real end-to-end check — SMTP handshake, catch-all detection, the works — and tells you exactly why an address is risky or undeliverable.

---

## How It Works

For every email address, `mailgo` runs a three-stage pipeline:

| Stage | Check | What It Catches |
|-------|-------|-----------------|
| 1 | **Syntax validation** | Malformed addresses |
| 2 | **SMTP handshake** | Mailboxes that don't exist |
| 3 | **Catch-all detection** | Servers that accept everything (unreliable) |

Every address gets one of four verdicts: `deliverable` · `undeliverable` · `risky` · `unknown`

---

## Quick Start

### macOS / Linux — Homebrew

```bash
brew tap Hrid-a/mailgo
brew install mailgo
```

### Windows — Scoop

```bash
scoop bucket add mailgo https://github.com/Hrid-a/scoop-mailgo.git
scoop install mailgo
```

### Go install

```bash
go install github.com/Hrid-a/mailgo@latest
```

### Pre-built binary

Download the archive for your platform from the [Releases page](https://github.com/Hrid-a/mailgo/releases), extract it, and place the binary on your `PATH`.

### Verify it works

```bash
mailgo verify user@example.com
```

---

## Usage

### Flags

| Flag | Description |
|------|-------------|
| `--json` | Output results as JSON instead of plain text |
| `--output`, `-o` | Write results to a file (e.g. `results.json`) |

### Single address

```bash
$ mailgo verify someone@company.com

Email       : someone@company.com
Domain      : company.com
Valid       : true
Status      : deliverable
Host Exists : true
Catch-All   : false
Deliverable : true
Full Inbox  : false
Disabled    : false
```

### JSON output

```bash
$ mailgo verify someone@company.com --json
```

```json
{
  "email": "someone@company.com",
  "domain": "company.com",
  "valid": true,
  "status": "deliverable",
  "smtp": {
    "host_exists": true,
    "catch_all": false,
    "deliverable": true,
    "full_inbox": false,
    "disabled": false
  }
}
```

### Bulk verification

Put one email per line in a `.txt` file, then:

```bash
$ mailgo verify --file emails.txt --output results.json
```

Results are written to `results.json`, one JSON object per address.

### Verdicts

| Verdict | Meaning |
|---------|---------|
| `deliverable` | Mailbox exists and should receive mail |
| `undeliverable` | Bad syntax or SMTP rejected the address |
| `risky` | SMTP connected but RCPT TO was rejected — may be anti-harvesting policy, not a dead address |
| `unknown` | Catch-all domain or SMTP unreachable — can't confirm either way |

---

## Contributing

### Clone the repo

```bash
git clone https://github.com/Hrid-a/mailgo.git
cd mailgo
```

### Build the binary

```bash
go build -o mailgo .
```

### Run the tests

```bash
go test ./...
```

### Submit a pull request

Fork the repository and open a pull request to the `main` branch. If you hit a mail server that produces a false positive or unexpected result, opening an issue with the domain (redacted if needed) is also very helpful.

---

## License

MIT
