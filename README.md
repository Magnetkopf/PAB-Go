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
