package enums

type Status string

const (
	StatusOK  Status = "OK"
	StatusNO  Status = "NO"
	StatusBAD Status = "BAD"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusOK, StatusNO, StatusBAD:
		return true
	default:
		return false
	}
}
