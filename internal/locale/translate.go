package locale

// T translates a piece of UI text. The English string is the key, so
// call sites stay readable and a string without a translation renders
// in English. Format strings are translated before formatting:
// fmt.Sprintf(locale.T("Loading %s..."), name).
func T(s string) string {
	if current == RU {
		if r, ok := ru[s]; ok {
			return r
		}
	}
	return s
}

// ru holds the Russian text by English key. The tables live in the
// ru_*.go files next to the screens they cover and register through
// add from init.
var ru = map[string]string{}

func add(table map[string]string) {
	for k, v := range table {
		ru[k] = v
	}
}
