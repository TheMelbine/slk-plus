package locale

import (
	"testing"
	"time"
)

func TestRussian(t *testing.T) {
	Set("ru")
	t.Cleanup(func() { Set("en") })

	thu := time.Date(2026, time.October, 15, 12, 43, 0, 0, time.UTC)
	cases := []struct{ got, want string }{
		{HeaderDate(thu, false), "Чт 15 окт, "},
		{HeaderDate(thu, true), "Чт 15 окт 2026, "},
		{SearchDate(thu, true), "15 окт 2026, "},
		{ShortDate(thu, false), "15 окт"},
		{DaySeparator(thu, 0), "Сегодня"},
		{DaySeparator(thu, 1), "Вчера"},
		{DaySeparator(thu, 3), "Четверг"},
		{DaySeparator(thu, 30), "Четверг, 15 октября 2026"},
		{FullDate(thu), "15 октября 2026"},
		{Replies(1), "1 ответ"},
		{Replies(3), "3 ответа"},
		{Replies(11), "11 ответов"},
		{Replies(21), "21 ответ"},
		{Ago(10 * time.Second), "только что"},
		{Ago(time.Minute), "1 минуту назад"},
		{Ago(2 * time.Hour), "2 часа назад"},
		{Ago(5 * 24 * time.Hour), "5 дней назад"},
		{Ago(400 * 24 * time.Hour), "1 год назад"},
		{Ago(5 * 365 * 24 * time.Hour), "5 лет назад"},
		{AgoShort(7 * time.Minute), "7м"},
		{AgoShort(3 * 24 * time.Hour), "3д"},
		{LastReply("только что"), "Последний ответ только что"},
		{LastBy("Alice"), "последний от Alice"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

func TestEnglishDefault(t *testing.T) {
	Set("klingon")
	if Current() != EN {
		t.Fatalf("unknown language should fall back to en, got %q", Current())
	}
	thu := time.Date(2026, time.October, 15, 12, 43, 0, 0, time.UTC)
	cases := []struct{ got, want string }{
		{HeaderDate(thu, false), "Thu Oct 15, "},
		{HeaderDate(thu, true), "Thu Oct 15 2026, "},
		{SearchDate(thu, false), "Oct 15, "},
		{ShortDate(thu, true), "Oct 15, 2026"},
		{DaySeparator(thu, 2), "Thursday"},
		{DaySeparator(thu, 9), "Thursday, October 15, 2026"},
		{Replies(1), "1 reply"},
		{Replies(2), "2 replies"},
		{Ago(time.Minute), "1 minute ago"},
		{Ago(3 * time.Hour), "3 hours ago"},
		{AgoShort(30 * time.Second), "now"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}
