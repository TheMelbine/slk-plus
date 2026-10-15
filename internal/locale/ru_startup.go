package locale

func init() {
	add(map[string]string{
		"Initializing...":           "Запуск...",
		"Connecting to ":            "Подключение к ",
		" (failed)":                 " (ошибка)",
		"Loading messages...":       "Загрузка сообщений...",
		"Loading older messages...": "Загрузка старых сообщений...",
		"No messages yet":           "Сообщений пока нет",
	})
}
