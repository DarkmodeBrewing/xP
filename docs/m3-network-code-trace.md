# M3.2 — Network lifecycle code trace (2026-10-24)

This inventory is based on direct inspection of the named source files on `master`. Findings are code-level risks, not yet reproduced failures.

## Startup and ownership

- `cmd/xd/xd.go:Run` creates storage, swarms, RPC, then launches one transport goroutine per swarm.
- When `conf.I2P.Disabled` is false, `runI2PFunc` creates an `i2p.Session`, registers it as a closer, repeatedly calls `Open`, passes it to `sw.ObtainedNetwork`, then calls `sw.Run`.
- When I2P is disabled and LokiNet is enabled, `runLokiNetFunc` instead uses `lib/network/inet.Session`, whose `Dial` uses `net.DialTCP`. **I2P-only is not an unconditional property of the executable.** This is an explicit configuration branch, not an observed automatic fallback.
- `lib/config/i2p.go:CreateSession` calls `i2p.NewSession`, creating a `samSession` with a buffered lookup channel.
- `lib/network/i2p/sam.go:Open` calls `OpenControlSocket` (TCP SAM HELLO), `keys.ensure`, `createSession("STREAM")`, starts `runLookups`, and requests `LookupI2P("ME")`.
- `OpenControlSocket` calls `net.Dial("tcp", s.addr)` and performs a line-based SAM handshake; no explicit dial or handshake deadline.
- Outgoing peer streams use `DialI2P` -> `OpenControlSocket` -> `STREAM CONNECT`; inbound streams use `i2pListener.Accept` -> `OpenControlSocket` -> `STREAM ACCEPT`.

## Prioritized findings

### P0: startup cleanup can block indefinitely

`samSession.Close` sends `nil` to `s.lookup` before closing the control socket. When `Open` fails before `go s.runLookups()`, there is no consumer; the send may fit the buffered channel once, but repeated closes or a full channel can block. More critically, `Close` is not an idempotent coordinated shutdown and can block when the lookup queue is full. Reproduce with an unstarted worker and full queue; do not assume the first send blocks (channel capacity is 18).

### P0: lookup worker failure can strand callers

`runLookups` exits on its first I/O error. `LookupI2P` sends a request and waits for an unbuffered reply without a timeout, cancellation, or worker-liveness check. Pending and subsequent requests can wait forever if the worker exits. Shutdown also risks racing requests and worker termination.

### P1: missing network deadlines

`OpenControlSocket` uses `net.Dial` and `readLine` without deadlines. `DialI2P`, `i2pListener.Accept`, `createSession`, and `Keyfile.ensure` also perform blocking protocol I/O without explicit timeouts. A stalled router can prevent prompt startup or shutdown.

### P1: SAM response/error handling

`OpenControlSocket` overwrites write errors with the subsequent `readLine` result. Its token parser can accept malformed replies if it sees `RESULT=OK` without validating the complete response. `readLine` reads one byte at a time and has no maximum line length; it indexes `buff[n-1]` without checking for a zero-byte read.

### P1: reconnect lifecycle is incomplete

In `cmd/xd/xd.go:runI2PFunc`, an error from `sw.Run` calls `sw.LostNetwork` and replaces the session, but does not explicitly close the old session first. An `Open` error replaces the session after `Open` attempts its own cleanup. No exponential backoff is visible on the `sw.Run` failure branch. `Context.Close` and transport goroutines do not use a shared cancellation context.

### P1: stream acceptance and ownership

`i2pListener.Close` only clears its local session reference; it does not cancel an in-progress `Accept`. `samSession.Accept` constructs a new listener on every call. Stream sockets have no explicit handshake deadline.

### P2: separate transport and configuration policy

`lib/network/inet/inet.go` uses normal TCP sockets for the optional LokiNet transport. The daemon chooses it only when I2P is disabled and LokiNet is enabled. This must be documented accurately: the current code does not prove an automatic clearnet fallback, but the executable is not strictly I2P-only in every configuration.

## Suggested next patch sequence

1. **M3.3**: Add deterministic tests reproducing lookup-worker failure, full-queue shutdown, and stalled SAM handshake.
2. **M3.4**: Introduce a session-owned cancellation/done mechanism and ensure pending lookups receive errors; make `Close` idempotent.
3. **M3.5**: Add bounded SAM dial/handshake deadlines, safe line-length limits, and strict reply handling.
4. **M3.6**: Fix session replacement cleanup and add bounded reconnect backoff; test repeated router restarts.
5. **M3.7**: Test egress policy and both router implementations. Explicitly distinguish LokiNet mode from I2P mode.

Do not change transport behavior until the relevant failing tests exist.
