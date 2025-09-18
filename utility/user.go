package utility

type Status int

const (
	StatusOK      Status = 0 // User is OK.
	StatusBan     Status = 1 // User is Ban.
	StatusTempBan Status = 2 // User is temporary banned.
	StatusDel     Status = 3 // User is deleted.
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusBan:
		return "Ban"
	case StatusDel:
		return "Del"
	case StatusTempBan:
		return "TempBan"
	default:
		return "Unknown"
	}
}
