package locale

// Message panes, thread panel, Activity, Threads, search, compose,
// reactions list and the image preview.
func init() {
	add(map[string]string{
		"Only visible to you":         "Видно только вам",
		"b to respond":                "b — ответить",
		"↗ open in Slack to interact": "↗ откройте в Slack, чтобы взаимодействовать",
		"Shared message":              "Пересланное сообщение",
		"Message from %s":             "Сообщение от %s",
		"[File]":                      "[Файл]",
		"[Image]":                     "[Картинка]",
		"(edited)":                    "(изменено)",
		"replied to a thread":         "ответил в треде",
		"-- more below --":            "-- ниже ещё --",
		"new":                         "новые",
	})
	add(map[string]string{
		"Button":    "Кнопка",
		"Select":    "Выбрать",
		"Pick date": "Выбрать дату",
		"Pick time": "Выбрать время",
		"Options":   "Варианты",
		"control":   "элемент",
		"image":     "картинка",
		"(no url)":  "(нет ссылки)",
		"loading…":  "загрузка…",
	})
	add(map[string]string{
		"esc close":          "esc — закрыть",
		"No thread selected": "Тред не выбран",
		"No replies yet":     "Ответов пока нет",
	})
	add(map[string]string{
		"Open Slack to see this reminder": "Напоминание видно в Slack",
		"no activity":                     "активности нет",
		"no unread activity":              "непрочитанной активности нет",
		"You: ":                           "Вы: ",
	})
	add(map[string]string{
		"Threads list unavailable — Slack subscription state could not be fetched. slk will retry on the next reconnect.": "Список тредов недоступен — Slack не отдал состояние подписок. slk попробует снова при следующем переподключении.",
		"no threads":          "тредов нет",
		"(parent not loaded)": "(сообщение ещё не загружено)",
	})
	add(map[string]string{
		"Type a query, Enter to search...": "Введите запрос, Enter — искать...",
		"Search Workspace":                 "Поиск по воркспейсу",
		"Searching…":                       "Ищем…",
		"No results":                       "Ничего не найдено",
		"showing %d of %d":                 "показано %d из %d",
	})
	add(map[string]string{
		"Message #%s... (i to insert)": "Сообщение в #%s... (i — писать)",
		"Reactions":                    "Реакции",
		"scroll":                       "листать",
	})
	add(map[string]string{
		"Esc/q close  •  Enter open in system viewer": "Esc/q — закрыть  •  Enter — открыть в просмотрщике",
		"Esc/q close":       "Esc/q — закрыть",
		"prev":              "назад",
		"next":              "вперёд",
		"%s  Loading %s...": "%s  Загрузка %s...",
		"%s  Loading...":    "%s  Загрузка...",
	})
}
