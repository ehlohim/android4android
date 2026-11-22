# M0 acceptance checks

M0 proves the Android app's media path, not that an internet calling API can answer a call.

## Gate 1: reference host and unattended boot

- Use an authorized Linux x86_64 host with usable KVM. Capture the host, Cuttlefish build and matching image identifiers in the test log.
- Run the diagnostic and retain its result. Passing the checks is a prerequisite, not proof the device boots.
- Boot without opening the viewer. Confirm the guest is online, launch the approved test app, and verify an expected UI state without a visible desktop.

## Gate 2: bidirectional guest audio

- Keep the viewer closed and do not route through a physical microphone or speaker.
- Inject a known synthetic fixture toward the guest microphone and independently capture a different known fixture from guest playback.
- Verify both directions run concurrently, distinct fixtures are observed on the expected sides, and stopping one session tears down its streams. Record only measurements and test IDs, not raw audio.
- Repeat inside the approved SIP app with a lab endpoint. An offline fixture run or a provider-only call does not pass this gate.

## Gate 3: Eros's acceptance test

After the lab run passes and both test endpoints are authorized, call the test number from a separate phone. Say the agreed test phrase and listen for the agent to respond. Keep the viewer closed throughout. Repeat in the outbound direction, then confirm both calls disconnected cleanly. Record pass, fail, or blocked and the first unexpected behavior. Test background ringing and restart recovery separately.

Do not put test phone numbers, authentication material, screenshots of account pages, or call data in public logs or issues. Real voice and messaging are opt-in and not part of default CI.
