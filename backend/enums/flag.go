package enums

type Flag string

const (
	FlagSeen     Flag = "\\Seen"
	FlagAnswered Flag = "\\Answered"
	FlagFlagged  Flag = "\\Flagged"
	FlagDeleted  Flag = "\\Deleted"
	FlagDraft    Flag = "\\Draft"
	FlagRecent   Flag = "\\Recent"
)

func (f Flag) IsValid() bool {
	switch f {
	case FlagSeen, FlagAnswered, FlagFlagged, FlagDeleted, FlagDraft, FlagRecent:
		return true
	default:
		return false
	}
}
