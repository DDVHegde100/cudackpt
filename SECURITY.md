# Security

## Supported versions

| Version | Supported |
|---------|-----------|
| 0.3.x   | Yes       |
| 0.2.x   | Yes       |
| 0.1.x   | Yes       |

## Reporting a vulnerability

Do **not** open a public GitHub issue for security-sensitive reports.

Email **ddvhegde100@gmail.com** with:

- Description of the issue and impact
- Steps to reproduce
- Affected version(s)
- Any suggested fix or mitigation

You should receive a response within a few business days.

## RPC authentication

When exposing the shim control socket beyond localhost, set a shared secret in `/etc/cudackpt.conf` (`rpc_secret=...`) or via environment:

```bash
export CUDACKPT_RPC_SECRET="$(openssl rand -hex 32)"
```

The shim target process must also receive `CUDACKPT_RPC_SECRET` at `LD_PRELOAD` time.

See [docs/OPERATIONS.md](docs/OPERATIONS.md) for deployment guidance.
