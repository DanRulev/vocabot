# 📚 `vocabot` — Telegram Bot for Learning English Words

> `vocabot` is a feature-rich Telegram bot designed to help users learn and memorize English vocabulary through interactive quizzes, daily word suggestions, and personalized progress tracking. Built with **Go**, **PostgreSQL**, and modern DevOps practices, it's a clean, scalable, and testable project perfect for language learners and developers alike.

---

## 🚀 Features

- ✅ **Daily New Word** — Discover a random English word with translation, pronunciation, examples, synonyms, and alternative translations.
- 🧠 **Interactive Quiz** — Test your knowledge: choose the correct translation from multiple options.
- 📊 **Progress Tracking** — View detailed statistics for learned words and quiz performance.
- 🗂 **Personal Vocabulary List** — Browse your known and unknown words with pagination.
- 🔁 **Interactive Menus & Inline Buttons** — Smooth UX with Telegram-native navigation.
- 💾 **In-Memory Caching** — Store active quizzes and word sessions to avoid duplication.
- 🌐 **External APIs** — Powered by:
  - [MyMemory](https://mymemory.translated.net/) – High-quality translation
  - [ftapi.pythonanywhere.com](https://ftapi.pythonanywhere.com/) – Dictionary definitions, examples, synonyms
  - [random-word-api.vercel.app](https://random-word-api.vercel.app/) – Random English words

---

## 🛠 Tech Stack

| Layer | Technology |
|------|------------|
| Language | Go 1.21+ |
| Framework | [`github.com/go-telegram-bot-api/telegram-bot-api/v5`](https://github.com/go-telegram-bot-api/telegram-bot-api) |
| Database | PostgreSQL |
| ORM | [`sqlx`](https://github.com/jmoiron/sqlx) |
| Config | [`viper`](https://github.com/spf13/viper) |
| Validation | [`validator`](https://github.com/go-playground/validator) |
| Logging | [`zap`](https://github.com/uber-go/zap) |
| CI/CD | GitHub Actions |
| Containerization | Docker + Docker Compose |
| Testing | `go test`, `golangci-lint`, `mockgen` |

---

## 🧰 Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/DanRulev/vocabot.git
cd vocabot
```

### 2. Environment Variables
Create a .env file in the root directory:

```env
CONFIG_NAME=default

BOT_TOKEN=

CONTAINER_NAME=

DB_HOST=db
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_SSL=
```
*🔐 ⚠️ Never commit your BOT_TOKEN to version control in production!*

### 3. Configuration

Create `configs/default.yaml:`

```yaml
app:
  timeout: 10s

env: production

db:
  conn:
    host: db
    port: "5432"
    user: botuser
    password: verysecretpass123
    name: vocabot
    ssl: disable
  cfg:
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_life_time: 1h
    conn_max_idle_time: 30m
```

### 4. Run with Docker

```bash
# Build and start services
make docker-build
make docker-up

# View logs
make docker-logs

# Stop
make docker-down
```

*This starts the bot and a PostgreSQL container.*

---

## 🧪 Testing & Linting

```bash
# Run unit tests
make test

# Check code formatting
gofmt -l .

# Run linter (via GitHub Actions or locally)
golangci-lint run
```

##### Generate Mocks (for testing)

```bash
make mock
```

Mocks are generated into:

- `internal/repository/mock/`
- `internal/service/mock/`
- `internal/bot/mock/`

## 🤖 Bot Commands & UI

| Command | Description |
|--------|-------------|
| `/start` | Show welcome screen and main menu |
| `/help` | Show help message |

### Main Menu Buttons

- **📚 New Word** — Get a new English word
- **🧠 Quiz** — Take a quiz
- **❗My Words** — View your vocabulary list
  - ✅ Learned words
  - ❌ Not learned words
- **📊 My Progress**
  - 📚 Word statistics
  - 🧠 Quiz statistics
- **ℹ️ Help** — Show help

All interactions are handled via buttons and inline callbacks.

## 🔄 CI/CD Pipeline

GitHub Actions workflow runs on every push (except `main`/`master`)

## 🛑 Disclaimer

- This bot relies on **third-party APIs** (MyMemory, ftapi.pythonanywhere.com, Vercel API), which may have **rate limits**, **downtime**, or change their terms of use without notice.
- The `BOT_TOKEN` included in the `.env` file is for **demonstration purposes only** and should **never be used in production**. Replace it with your own token from [@BotFather](https://t.me/BotFather).
- This project is intended for **educational and personal use**. Do not use it for commercial purposes without proper testing, attribution, and compliance with API terms.

## 📄 License

This project is open-source. Please add a license file (`LICENSE`) to specify the terms of use (e.g., MIT, Apache 2.0, GPL). Without an explicit license, the code is **copyrighted** and not free to use or distribute.

## 🙌 Contributing

Contributions are **welcome and appreciated**! Whether you want to fix bugs, improve documentation, add features, or report issues — your help makes this project better.

### How to Contribute:
1. Fork the repository
2. Create a new feature branch
3. Commit your changes
4. Push to your branch
5. Open a Pull Request

Also feel free to:
- 🐛 Report **bugs** via Issues
- 💡 Suggest **new features or improvements**
- 📚 Improve **documentation**
