package host

// History is every line the message scroll showed, oldest first (F5 shows it).
var History []string

// The history keeps at most historyMax lines, dropping the older half when full.
const historyMax = 2000

// AddHistory appends a finished line of the message scroll.
func AddHistory(s string) {
	History = append(History, s)
	if len(History) > historyMax {
		History = append([]string(nil), History[len(History)-historyMax/2:]...)
	}
}
