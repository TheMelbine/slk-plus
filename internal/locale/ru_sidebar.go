package locale

// Sidebar, pickers, profile and status menus, help chrome, dialogs.
func init() {
	add(map[string]string{
		// sidebar
		"Channels":        "Каналы",
		"Direct Messages": "Личные сообщения",
		"Apps":            "Приложения",
		"Starred":         "Избранное",
		"(unnamed)":       "(без имени)",
		"Threads":         "Треды",
		"Activity":        "Активность",
		"No channels":     "Нет каналов",

		// pickers and modals
		"Type to filter...":                     "Введите для фильтра...",
		"Type to filter…":                       "Введите для фильтра…",
		"Switch Channel":                        "Переход в канал",
		"Forward message to…":                   "Переслать сообщение в…",
		"No matching channels":                  "Нет подходящих каналов",
		"Switch Workspace":                      "Смена воркспейса",
		"No matching workspaces":                "Нет подходящих воркспейсов",
		"Switch Theme":                          "Смена темы",
		"No matching themes":                    "Нет подходящих тем",
		"Add Reaction":                          "Добавить реакцию",
		"No matching emoji":                     "Нет подходящих эмодзи",
		"New message":                           "Новое сообщение",
		"No users available":                    "Нет доступных пользователей",
		"No users match %q":                     "Никто не подходит под %q",
		"type a name…":                          "введите имя…",
		" [ext]":                                " [внешний]",
		"space toggle  enter open  esc cancel":  "space — выбрать  enter — открыть  esc — отмена",
		"  MPIM limit reached":                  "  Достигнут лимит участников",
		"j/k move   enter select   esc/q close": "j/k — перемещение   enter — выбрать   esc/q — закрыть",
		" (not in channel)":                     " (не в канале)",
		" (ext)":                                " (внешний)",

		// help overlay
		"Keybindings":             "Клавиши",
		"No matching keybindings": "Нет подходящих клавиш",
		"filter: ":                "фильтр: ",
		"Press / to search":       "Нажмите / для поиска",
		"/ search   esc/q close":  "/ — поиск   esc/q — закрыть",

		// presence menu
		"Status":                          "Статус",
		"Active":                          "В сети",
		"Away":                            "Нет на месте",
		"Snooze for 20 minutes":           "Не беспокоить 20 минут",
		"Snooze for 1 hour":               "Не беспокоить 1 час",
		"Snooze for 2 hours":              "Не беспокоить 2 часа",
		"Snooze for 4 hours":              "Не беспокоить 4 часа",
		"Snooze for 8 hours":              "Не беспокоить 8 часов",
		"Snooze for 24 hours":             "Не беспокоить 24 часа",
		"Snooze until tomorrow morning":   "Не беспокоить до завтрашнего утра",
		"Snooze custom...":                "Не беспокоить: своё время...",
		"End snooze / DND":                "Снять «не беспокоить»",
		"No matching options":             "Нет подходящих вариантов",
		"Snooze for how many minutes?":    "На сколько минут не беспокоить?",
		"Enter to snooze · Esc to cancel": "Enter — включить · Esc — отмена",

		// user profile and peer status
		"Profile":                               "Профиль",
		"K / esc / q close":                     "K / esc / q — закрыть",
		"e copy email · K / esc / q close":      "e — скопировать почту · K / esc / q — закрыть",
		"Loading profile…":                      "Загрузка профиля…",
		"external":                              "внешний",
		"APP":                                   "ПРИЛОЖЕНИЕ",
		"deactivated":                           "деактивирован",
		"In a huddle":                           "В хадле",
		"Do not disturb":                        "Не беспокоить",
		" until ":                               " до ",
		"Local time":                            "Местное время",
		"Email":                                 "Почта",
		"Phone":                                 "Телефон",
		"same time as you":                      "как у вас",
		" from you":                             " от вас",
		"Rate limited — try again in %ds":       "Слишком много запросов — повторите через %d с",
		"Couldn't load full profile: timed out": "Не удалось загрузить профиль: таймаут",
		"Couldn't load full profile: ":          "Не удалось загрузить профиль: ",

		// inline attachments
		"Loading %s...": "Загрузка %s...",
		"[File]":        "[Файл]",
		"[Image]":       "[Картинка]",

		// confirm prompt
		"confirm": "подтвердить",
		"cancel":  "отмена",
	})
}
