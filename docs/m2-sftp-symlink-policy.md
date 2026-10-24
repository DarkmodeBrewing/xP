# M2.4 — SFTP recursive deletion policy

## Policy

Recursive removal must not traverse symbolic links. A symlink encountered as a child is removed as a link, not followed into its target. A symlink supplied as the top-level removal path is likewise unlinked rather than recursively traversed.

Use SFTP `Lstat` (not `Stat`) both for the top-level path and for each child immediately before deciding whether to recurse. This avoids following stable symlinks and catches entries changed to symlinks after the directory listing.

## Threat boundary and remaining risk

This does **not** provide the same containment guarantees as local `os.Root`. SFTP is path-based, and an actor with write access to parent directories can replace a directory with a symlink between `Lstat` and a subsequent operation. The remote server ultimately resolves paths, and may have implementation-specific behavior.

Do not claim protection against concurrent hostile remote filesystem mutation. A server-side confinement mechanism (such as chroot or dedicated restricted account) is required for that stronger guarantee.

## Follow-up integration coverage

Use a local in-process SFTP server fixture to verify:
- A symlink nested under a torrent directory is unlinked without affecting the target.
- A top-level symlink is unlinked without affecting the target.
- Nested ordinary directories are removed recursively.
- Permission and server failures abort deletion without traversing unrelated paths.
- A hostile concurrent directory replacement is documented as an unsupported trust boundary unless server-side confinement is available.
