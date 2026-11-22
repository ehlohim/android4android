# Security

## Prototype status

There are no supported production releases or security guarantees yet. Treat this project as experimental software for isolated test systems, not as a hardened multi-tenant Android service.

## Reporting a vulnerability

Do not put exploit details, credentials, personal data, or private recordings in public issues or pull requests.

During private development, report concerns directly to Eros through the existing private project conversation. Once a GitHub repository is published, use its **Security → Report a vulnerability** option if private vulnerability reporting is enabled. If no private channel is available, ask the maintainer how to report privately without disclosing the vulnerability. No response-time commitment is established yet.

Include a sanitized reproduction, affected revision, impact, and any safe mitigation. Use dummy accounts and synthetic data.

## Safe development boundaries

- Keep ADB, control endpoints, consoles, and guest displays off public networks. Authentication and access control must be reviewed before exposing them.
- Use disposable test guests and least-privileged host access. Access to KVM or container/host administration is a security boundary, not an ordinary test prerequisite.
- Never extract another application's credentials or commit secrets, session data, device snapshots, phone numbers, or user audio.
- Do not record, retain, transmit, call, or message real people without authorization. Provider charges and consent must be settled before live testing.
- Keep ordinary tests offline and credential-free. Run Android/KVM and telephony integration separately on an authorized lab host.
- Review changes to workflows carefully. Do not execute untrusted contributions on a privileged or credential-bearing runner.
- If a secret is exposed, revoke or rotate it through its issuing service and notify the owner privately. Removing the current file alone does not remove history or invalidate the secret.

Google device certification, carrier IMS/eSIM support, and production isolation remain separate workstreams. Their presence must not be inferred from a successful unit test or demo.
