package grove

// Register records one or more repos (name → path) in the GitGroove metadata.
//
// Requirements:
//   - rootAbsPath must point to a valid Git repository already initialized with GitGroove (.gg exists).
//   - Each repo entry uses the repo's unique name as the key and its directory path as the value.
//   - All names must be globally unique. If any name in the batch is already registered, the entire
//     registration step is aborted with no changes applied.
//   - All paths must exist, be directories, and reside within the Git project root.
//   - Repos cannot contain their own .git directory.
//   - Updating an existing repo's path is not supported here and must be handled by a dedicated command.
//
// Behavior:
//   - Valid repo entries are appended to .gg/repos.jsonl.
//   - No file is overwritten; the metadata is strictly append-only.
//   - After updating metadata, the changes are committed to the gitgroove/system branch.
//   - Registration does not define hierarchy—linking repos is handled separately via the link command.
//
// Register must be atomic: if any repo fails validation, nothing is written or committed.
func Register(rootAbsPath string, repos map[string]string) {
}
