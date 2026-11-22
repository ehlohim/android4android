# android4android

A self-hosted Android workspace for AI agents. The goal is to give an agent a persistent Android device it can use without an operator watching, including ordinary apps and real voice and messaging through supported providers.

## Status

Early feasibility work. The host diagnostic checks prerequisites for the Linux Cuttlefish reference target; it does not boot Android, install dependencies, or verify a call. Local audio fixtures test the harness, not guest audio or a live phone service. No Google account, carrier eSIM, or device certification is required by the initial design.

## Local checks

With Go installed, from the project directory:

```sh
go test ./...
go vet ./...
go run ./cmd/android4android doctor --json
```

A built diagnostic binary exits 2 when prerequisites are missing. A reference host needs Linux x86_64 with usable KVM and a compatible Cuttlefish host package and Android image. Windows/macOS local runtimes are not qualified.

## First milestone

1. Reproduce a headless Cuttlefish boot on a qualified Linux host.
2. With an approved test APK, verify simultaneous software microphone and speaker paths with no viewer attached.
3. Repeat with a separately installed SIP calling app on a controlled lab session.
4. With authorized test endpoints, verify real inbound and outbound calls through the Android app. Provider-only audio is not a substitute.

No calls, texts, accounts, credentials, or user audio are needed for the default tests. Ask before spending money or using a real number. Carrier Wi-Fi Calling/eSIM and Google certification are separate research tracks.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).
