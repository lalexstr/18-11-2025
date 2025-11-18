Используемые технологии
• Go
• Gin (HTTP сервер)
• Gorm (ORM)
• SQLite (локальная база данных)
• gofpdf (генерация PDF файлов)

cmd/server/main.go точка входа
internal/handlers обработчики HTTP запросов
internal/models модели Gorm
internal/worker фоновая очередь для проверки ссылок
internal/storage подключение SQLite и миграции
internal/pdf генерация PDF отчета

Как работает проверка ссылок
1. При запросе POST /links создается LinkSet.
2. Каждая ссылка сохраняется в таблицу Link со статусом “pending” и Processed=false.
3. ID созданной ссылки помещается в очередь в воркере.
4. Воркер делает HTTP запрос к ссылке.
5. Если код ответа от 200 до 399 — статус ok, иначе fail.
6. Статус сохраняется в базу, Processed становится true.

При перезапуске приложения все ссылки, у которых Processed=false, снова добавляются в очередь.

Эндпоинты:

POST /links

JSON
`
{
  "links": [
    "https://example.com",
    "http://google.com"
  ]
}
`

GET /report?links_num=1
Возвращает PDF файл с результатами проверки.

Можно указать несколько наборов:
GET /report?links_num=1,2,3

GET /ping
Простой ответ для проверки работоспособности.
