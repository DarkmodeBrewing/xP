# M2.3 — Filesystem driver security review

## Scope and current status

M2.2 rejects malicious torrent metadata path components. This does **not** make
filesystem operations symlink-safe.

### Local driver (`lib/fs/std.go`)

- `OpenFileReadOnly` uses `os.Open`; `OpenFileWriteOnly` uses `os.OpenFile`.
  Both can follow symlinks in the final component **and parent directories**.
- `EnsureDir` and `EnsureFile` delegate to `lib/util`; audit their
  symlink and directory-creation behavior before claiming confinement.
- `Move` creates destination parents and uses `os.Rename`. A symlink in a
  destination parent can redirect a move outside the expected tree.
- `RemoveAll` uses `os.RemoveAll`. Go does not recursively traverse a
  symlink as a directory, but a parent path may still resolve through a link.

**Required follow-up:** introduce a root-scoped filesystem API for torrent data
operations, using Go's `os.OpenRoot` / `os.Root` on supported platforms and
carefully designed relative paths. The existing `fs.Driver` API accepts
unrestricted absolute paths and cannot itself guarantee containment. Do not
attempt to solve races with `EvalSymlinks` followed by `Open`.

### SFTP driver (`lib/fs/sftp.go`)

- `Stat`, `Open`, `Create`, `Rename`, and directory creation operate on
  server-resolved paths. A remote server may follow symlinks; Go's local
  `os.Root` guarantees do not apply to SFTP.
- `removeAllDir` previously recursed using a bare child name instead of
  `Join(root, child)`; corrected in this PR.
- `RemoveAll` and `removeAllDir` still need explicit symlink policy,
  remote-server integration tests, and careful handling of changing entries.
- `Join` can panic on connection errors; investigate separately.

## Threat model and next steps

Treat torrent metadata as untrusted. The user-configured storage roots and
the filesystem itself may contain preexisting symlinks. A local actor changing
the directory tree concurrently is a stronger threat than malicious metadata
alone. SFTP remote servers are a separate trust boundary.

1. Refactor torrent data operations around root-relative, scoped handles.
2. Add tests with a symlink from a download root into a sibling directory,
   covering allocate, read, write, move, and delete.
3. Add concurrency tests that replace parent directories with symlinks.
4. Specify and test the SFTP symlink policy with a real SFTP server fixture.
5. Keep metadata/bitfield paths and explicitly configured destination roots
   distinct from torrent-controlled relative paths.

**Security claim for this PR:** SFTP recursive removal child-path correctness
and documented driver trust boundaries. **Not** full symlink confinement.
