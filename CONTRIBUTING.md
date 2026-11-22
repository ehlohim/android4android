# Contributing

android4android is an early feasibility prototype. Turing is the engineer; Eros owns product decisions and QA acceptance. This starter is being prepared privately. Repository publication and licensing are separate owner decisions.

## Keep changes reviewable

- Agree on the problem and acceptance criteria before a substantial change.
- Prefer small changes with tests and reproducible commands.
- Keep the Go scaffold standard-library-only unless a dependency is explicitly agreed.
- Separate observed results from proposed capabilities. A passing fixture is not proof of Android audio, an app call, or carrier support.
- Use synthetic fixtures. Never commit credentials, personal phone numbers, account exports, recordings, or device images.
- Do not add privileged automation, paid services, live calls, or account setup as a side effect of normal tests.

## Local checks

Use the Go version specified in `go.mod`, then run from the repository root:

```sh
gofmt -w cmd internal
go test ./...
go vet ./...
```

Review the diff after formatting. Tests must not require Android, KVM, a provider account, network credentials, or a real telephone endpoint.

## Two validation tracks

**Hosted CI:** formatting, Go unit tests, and static checks on an ordinary GitHub-hosted Linux runner. These checks validate only the scaffold and fixtures.

**Integration qualification:** separately authorized runs on an isolated Linux host with usable KVM and compatible Android/Cuttlefish components. Record the host/image versions, commands, expected and observed results, and sanitized evidence. Test real guest microphone/speaker paths without a viewer before claiming voice support. Live calling requires approved test endpoints, consent, and a cost limit. Do not attach a privileged lab runner to untrusted pull requests.

Eros reviews milestone evidence and accepts the result. Passing hosted CI alone does not qualify a runtime, app, or calling provider.

## Pull requests and issues

Include the change's purpose, scope, commands run, results, and remaining limitations. Flag changes to authentication, network exposure, device permissions, audio capture, storage, or provider use. For suspected vulnerabilities, follow [SECURITY.md](SECURITY.md), not a public issue.
