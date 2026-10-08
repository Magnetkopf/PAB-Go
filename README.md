# PAB-Go

[ayyyyano/Personal-AskBox](https://github.com/ayyyyano/Personal-AskBox), but Vue and Go

- 🌟Light weight
- 📚No database
- 💻Any os, any arch
- 🔠i18n

# Build

```
cd frontend
bun i
cd ../
make build
```

# Usage

## init

```sh
chmod +x ./pab-go
# setup admin info, then enjoying.
./pab-go
```

Use `--bind ip:port` to override the default `127.0.0.1:7212`, for example:

```sh
./pab-go --bind 0.0.0.0:7212
```

## Telegram bot

Open **Admin → Telegram bot** to enter a bot token and enable either feature:

- **New question notifications:** enter your Telegram user ID to receive new questions from both the website and the bot.
- **Questions from Telegram:** the bot accepts English-language text interactions in private chats. Set a total daily question limit (reset at 00:00 UTC) and choose whether one user may have multiple unanswered questions. Answer notifications to the asker are not implemented yet.

Telegram uses long polling, so run only one PAB-Go process for a bot token and remove any webhook configured for that bot. Telegram settings are stored in `data/telegram.json`; question-to-user mappings are stored separately in `data/telegram_questions.json`. A user is identified by the hex SHA-256 digest of `"T"` followed by the decimal Telegram user ID. This digest can be checked against guessed user IDs. Existing HMAC mappings cannot be converted to this format without the original user IDs; the old `telegram_identity.key` is no longer used. Messages sent while questions are disabled are ignored after the feature is re-enabled.

## systemd

`/etc/systemd/system/pab-go.service`

```toml
[Unit]
Description=PAB-go
After=network.target

[Service]
WorkingDirectory=/opt/pab-go
ExecStart=/opt/pab-go/pab-go --bind 0.0.0.0:7212
Restart=unless-stopped
// recommended to use a non-root user
User=root

[Install]
WantedBy=multi-user.target
```
