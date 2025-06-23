package enums

type FolderType string

const (
	FolderInbox  FolderType = "inbox"
	FolderSent   FolderType = "sent"
	FolderTrash  FolderType = "trash"
	FolderSpam   FolderType = "spam"
	FolderDrafts FolderType = "drafts"
)

func (f FolderType) IsValid() bool {
	switch f {
	case FolderInbox, FolderSent, FolderTrash, FolderSpam, FolderDrafts:
		return true
	default:
		return false
	}
}
