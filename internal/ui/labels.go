package ui

// All user-visible strings are defined here so they can be swapped out for
// translated versions in one place. Format strings (containing %s, %q, %d)
// are passed to fmt.Sprintf at the call site.

// ── Panel titles ──────────────────────────────────────────────────────────────

const (
	TitleBranches = "Branches"
	TitleCommits  = "Commits"
	TitleDetail   = "Detail"
)

// ── Branch panel ──────────────────────────────────────────────────────────────

const (
	SectionLocal         = "LOCAL"
	SectionRemote        = "REMOTE"
	MarkerCurrentBranch  = "▶ "
	MarkerUpstreamArrow  = " -> "
	FilterPrompt         = "  / filter"
	FilterActiveMark     = "  ~"
	PlaceholderBranches  = "filter branches..."
	NewBranchButtonLabel = "» new branch"
)

// ── Branch picker dialog ───────────────────────────────────────────────────────

const (
	BranchPickerMergeFmt          = "Merge %s into:"
	BranchPickerRebaseFmt         = "Rebase %s onto:"
	BranchPickerEmpty             = "(no matches)"
	BranchPickerFilterPlaceholder = "filter branches..."
)

// ── Commits panel ─────────────────────────────────────────────────────────────

const (
	PlaceholderCommits   = "hash, message, author..."
	MultiSelectHintLabel = "  [multi-select: space=toggle, Enter=squash]"
)

// ── Detail panel ──────────────────────────────────────────────────────────────

const (
	DetailLabelCommit = "commit "
	DetailLabelAuthor = "Author: "
	DetailLabelDate   = "Date:   "
	DetailLabelTags   = "Tags:   "
	DetailLabelFiles  = "Changed files:"
)

// ── Context menu ──────────────────────────────────────────────────────────────

const (
	MenuTitle          = "Actions"
	MenuCheckoutRemote = "Checkout (create local tracking)"
	MenuFetchRemote    = "Fetch remote"
	MenuCheckout       = "Checkout"
	MenuMerge          = "Merge into …"
	MenuRebase         = "Rebase onto …"
	MenuPush           = "Push to remote"
	MenuForcePush      = "Force-push to remote"
	MenuPullFromFmt    = "Pull from %s"
	MenuDeleteBranch   = "Delete branch"
	MenuCherryPick     = "Cherry-pick onto current branch"
	MenuRevert         = "Revert (create revert commit)"
	MenuCopyHash       = "Copy commit hash"
	MenuDropCommit     = "Drop commit from history"
	MenuAmend          = "Amend commit message"
	MenuSquash         = "Squash with next commit"
)

// ── Confirm dialogs ───────────────────────────────────────────────────────────

const (
	ConfirmDeleteBranch  = "Delete branch?"
	ConfirmForcePushFmt  = "Force-push %q to origin? This overwrites remote history."
	ConfirmDropCommitFmt = "Drop commit %s from history?"
)

// ── Amend dialog ──────────────────────────────────────────────────────────────

const (
	AmendTitle       = "Amend commit message:"
	AmendPlaceholder = "Commit message..."
)

// ── New branch dialog ─────────────────────────────────────────────────────────

const (
	NewBranchTitleFmt     = "New branch from %s:"
	NewBranchPlaceholder  = "branch-name"
	NewBranchErrEmpty     = "enter a branch name"
	NewBranchErrExistsFmt = "%q already exists"
)

// ── Dialog action hints ───────────────────────────────────────────────────────

const (
	HintSave    = "save"
	HintCreate  = "create"
	HintConfirm = "confirm"
	HintCancel  = "cancel"
	HintApply   = "apply"
	HintClear   = "clear"
)

// ── Status bar hint descriptions ──────────────────────────────────────────────

const (
	HintNextPanel       = "next panel"
	HintNavigate        = "navigate"
	HintMenu            = "menu"
	HintCheckout        = "checkout"
	HintFetch           = "fetch"
	HintMerge           = "merge"
	HintRebase          = "rebase"
	HintPush            = "push"
	HintPull            = "pull"
	HintDelete          = "delete"
	HintNewBranch       = "new branch"
	HintFilter          = "filter"
	HintScroll          = "scroll"
	HintMultiSelect     = "multi-select"
	HintExitMultiSelect = "exit multi-select"
	HintToggle          = "toggle"
	HintSquash          = "squash"
	HintAmend           = "amend"
	HintDrop            = "drop"
	HintCherryPick      = "cherry-pick"
	HintRevert          = "revert"
	HintQuit            = "quit"
)

// ── Status bar output prefixes ────────────────────────────────────────────────

const (
	StatusOKPrefix  = "✓ "
	StatusErrPrefix = "✗ "
)

// ── App-level error messages ──────────────────────────────────────────────────

const (
	ErrLoadCommitsFmt = "load commits: %s"
	ErrLoadDetailFmt  = "load detail: %s"
	ErrSquashTooFew   = "select at least 2 commits with [s] then [space]"
)
