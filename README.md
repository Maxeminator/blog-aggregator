Gator — это CLI-приложение для агрегации RSS-лент. Оно позволяет регистрировать пользователей, подписываться на фиды, собирать посты, и просматривать их в терминале. Работает на Go и использует PostgreSQL как базу данных.

Для запуска Gator тебе нужно установить Go (1.20+) и PostgreSQL (12+).

Установка:
  go install github.com/Maxeminator/blog-aggregator@latest

После установки бинарник будет доступен как gator, если $GOPATH/bin добавлен в PATH.

Настройка: необходимо создать файл ~/.gatorconfig.json следующего содержания:

{
  "current_user": ""
}

Также необходимо создать базу данных PostgreSQL (например, gator), задать URL подключения и применить миграции:

  goose -dir sql/schema postgres "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable" up

Для разработки можно использовать:

  go run .

Для сборки:

  go build -o gator .

Команды CLI:

- register <username> — регистрирует нового пользователя и сохраняет его в конфиг
- login <username> — авторизация под существующим пользователем
- addfeed <name> <url> — добавляет RSS-ленту и сразу подписывает на неё
- follow <url> — подписаться на уже добавленную RSS-ленту по URL
- unfollow <url> — отписаться от ленты
- feeds — список всех лент
- following — список лент, на которые подписан пользователь
- browse [limit] — посмотреть последние посты (по умолчанию limit = 2)
- agg <duration> — запускает бесконечный сборщик фидов с указанным интервалом (например, 30s или 1m)
- reset — удаляет всех пользователей (используется только для сброса/отладки)

Пример использования:
- gator register alice
- gator login alice
- gator addfeed "TechCrunch" https://techcrunch.com/feed/
- gator follow https://techcrunch.com/feed/
- gator agg 1m
- gator browse 5

Команда agg запускает фоновый бесконечный цикл сбора фидов. Она не должна DOS-ить источники. Используй разумные интервалы, например, 1m или больше. Остановить выполнение можно через Ctrl+C.

Посты сохраняются в базу данных и ассоциируются с фидом. Повторно сохранять один и тот же пост не получится — дубли по URL игнорируются.

Репозиторий на GitHub: https://github.com/Maxeminator/blog-aggregator
