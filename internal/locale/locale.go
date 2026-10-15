// Package locale holds the display language of the TUI's generated
// text: date labels, reply counts, relative times and the few words
// around them. It is package-level state like the theme in styles,
// seeded once at startup from appearance.language, because every
// render path must agree. Only "en" and "ru" are known; anything else
// is English. No I/O here.
package locale

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Lang is a supported display language.
type Lang string

const (
	EN Lang = "en"
	RU Lang = "ru"
)

var current = EN

// Set picks the display language. Unknown values fall back to English.
func Set(lang string) {
	if Lang(strings.ToLower(lang)) == RU {
		current = RU
		return
	}
	current = EN
}

// Current returns the active language.
func Current() Lang { return current }

var (
	ruWeekdays      = [...]string{"Воскресенье", "Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота"}
	ruWeekdaysShort = [...]string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
	// Genitive: the only case a day-of-month date uses.
	ruMonths      = [...]string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
	ruMonthsShort = [...]string{"янв", "фев", "мар", "апр", "мая", "июн", "июл", "авг", "сен", "окт", "ноя", "дек"}
)

// plural picks the Russian form for n: 1 → one, 2–4 → few, else many
// (with the 11–14 exception).
func plural(n int, one, few, many string) string {
	n10, n100 := n%10, n%100
	switch {
	case n10 == 1 && n100 != 11:
		return one
	case n10 >= 2 && n10 <= 4 && (n100 < 12 || n100 > 14):
		return few
	default:
		return many
	}
}

// HeaderDate is the date prefix of a selected message's header, with
// a trailing ", " so the clock follows: "Mon Jan 2, " or "Mon Jan 2
// 2006, " (withYear); "Пн 2 янв, " or "Пн 2 янв 2006, " in Russian.
func HeaderDate(t time.Time, withYear bool) string {
	if current == RU {
		s := ruWeekdaysShort[t.Weekday()] + " " + strconv.Itoa(t.Day()) + " " + ruMonthsShort[t.Month()-1]
		if withYear {
			s += " " + strconv.Itoa(t.Year())
		}
		return s + ", "
	}
	if withYear {
		return t.Format("Mon Jan 2 2006, ")
	}
	return t.Format("Mon Jan 2, ")
}

// SearchDate is the date prefix of a search-result timestamp, with a
// trailing ", ": "Jan 2, " or "Jan 2 2006, "; "2 янв, " or "2 янв
// 2006, " in Russian.
func SearchDate(t time.Time, withYear bool) string {
	if current == RU {
		s := strconv.Itoa(t.Day()) + " " + ruMonthsShort[t.Month()-1]
		if withYear {
			s += " " + strconv.Itoa(t.Year())
		}
		return s + ", "
	}
	if withYear {
		return t.Format("Jan 2 2006, ")
	}
	return t.Format("Jan 2, ")
}

// ShortDate is the compact date of a message header: "Jan 2" or "Jan
// 2, 2006" (withYear); "2 янв" or "2 янв 2006" in Russian.
func ShortDate(t time.Time, withYear bool) string {
	if current == RU {
		s := strconv.Itoa(t.Day()) + " " + ruMonthsShort[t.Month()-1]
		if withYear {
			s += " " + strconv.Itoa(t.Year())
		}
		return s
	}
	if withYear {
		return t.Format("Jan 2, 2006")
	}
	return t.Format("Jan 2")
}

// DaySeparator labels a day-divider row for the day t, daysAgo days
// before today: "Today", "Yesterday", a weekday name within the last
// week, else the full date ("Monday, January 2, 2006"; "Понедельник,
// 2 января 2006").
func DaySeparator(t time.Time, daysAgo int) string {
	switch {
	case daysAgo <= 0:
		if current == RU {
			return "Сегодня"
		}
		return "Today"
	case daysAgo == 1:
		if current == RU {
			return "Вчера"
		}
		return "Yesterday"
	case daysAgo < 7:
		if current == RU {
			return ruWeekdays[t.Weekday()]
		}
		return t.Format("Monday")
	}
	return LongDate(t)
}

// LongDate is the fully qualified date: "Monday, January 2, 2006";
// "Понедельник, 2 января 2006" in Russian.
func LongDate(t time.Time) string {
	if current == RU {
		return ruWeekdays[t.Weekday()] + ", " + FullDate(t)
	}
	return t.Format("Monday, January 2, 2006")
}

// FullDate is the date with the month spelled out: "January 2, 2006";
// "2 января 2006" in Russian.
func FullDate(t time.Time) string {
	if current == RU {
		return strconv.Itoa(t.Day()) + " " + ruMonths[t.Month()-1] + " " + strconv.Itoa(t.Year())
	}
	return t.Format("January 2, 2006")
}

// Replies counts thread replies: "1 reply" / "3 replies"; "1 ответ" /
// "3 ответа" / "5 ответов".
func Replies(n int) string {
	if current == RU {
		return strconv.Itoa(n) + " " + plural(n, "ответ", "ответа", "ответов")
	}
	if n == 1 {
		return "1 reply"
	}
	return strconv.Itoa(n) + " replies"
}

// Ago says how long d is, in its largest whole unit: "just now", "5
// minutes ago", "2 hours ago", "3 days ago", "2 months ago", "1 year
// ago"; "только что", "5 минут назад" in Russian.
func Ago(d time.Duration) string {
	if d < time.Minute {
		if current == RU {
			return "только что"
		}
		return "just now"
	}
	n, unit := split(d)
	if current == RU {
		forms := map[string][3]string{
			"minute": {"минуту", "минуты", "минут"},
			"hour":   {"час", "часа", "часов"},
			"day":    {"день", "дня", "дней"},
			"month":  {"месяц", "месяца", "месяцев"},
			"year":   {"год", "года", "лет"},
		}[unit]
		return fmt.Sprintf("%d %s назад", n, plural(n, forms[0], forms[1], forms[2]))
	}
	if n != 1 {
		unit += "s"
	}
	return fmt.Sprintf("%d %s ago", n, unit)
}

// split breaks d (at least a minute) into a count and an English unit
// name.
func split(d time.Duration) (int, string) {
	switch {
	case d < time.Hour:
		return int(d / time.Minute), "minute"
	case d < 24*time.Hour:
		return int(d / time.Hour), "hour"
	case d < 30*24*time.Hour:
		return int(d / (24 * time.Hour)), "day"
	case d < 365*24*time.Hour:
		return int(d / (30 * 24 * time.Hour)), "month"
	default:
		return int(d / (365 * 24 * time.Hour)), "year"
	}
}

// AgoShort is the coarse card time: "now" / "7m" / "3h" / "2d";
// "сейчас" / "7м" / "3ч" / "2д" in Russian.
func AgoShort(d time.Duration) string {
	ru := current == RU
	switch {
	case d < time.Minute:
		if ru {
			return "сейчас"
		}
		return "now"
	case d < time.Hour:
		return strconv.Itoa(int(d/time.Minute)) + pick(ru, "м", "m")
	case d < 24*time.Hour:
		return strconv.Itoa(int(d/time.Hour)) + pick(ru, "ч", "h")
	default:
		return strconv.Itoa(int(d/(24*time.Hour))) + pick(ru, "д", "d")
	}
}

func pick(ru bool, r, e string) string {
	if ru {
		return r
	}
	return e
}

// LastReply prefixes the thread line's age: "Last reply 2 hours ago".
func LastReply(when string) string {
	return pick(current == RU, "Последний ответ ", "Last reply ") + when
}

// LastBy prefixes the Threads card's last replier: "last by Alice".
func LastBy(name string) string {
	return pick(current == RU, "последний от ", "last by ") + name
}

// Thread is the breadcrumb's leading word.
func Thread() string { return pick(current == RU, "Тред", "Thread") }

// From is the breadcrumb's " from <author>" prefix.
func From() string { return pick(current == RU, " от ", " from ") }

// In joins an Activity label to its conversation: "Reminder in #ch".
func In() string { return pick(current == RU, " в ", " in ") }

// Reminder labels a saved-reminder Activity item.
func Reminder() string { return pick(current == RU, "Напоминание", "Reminder") }

// UnknownActivity labels an Activity item with no type.
func UnknownActivity() string {
	return pick(current == RU, "Неизвестное событие", "Unknown activity")
}
