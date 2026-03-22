# Развёртывание и администрирование SSH Arena

Это руководство предназначено для системных администраторов и DevOps‑инженеров, которые разворачивают и обслуживают сервер SSH Arena в production‑среде.

## Содержание

1. [Архитектура системы](#архитектура-системы)
2. [Требования к инфраструктуре](#требования-к-инфраструктуре)
3. [Настройка SSH‑аутентификации](#настройка-ssh‑аутентификации)
4. [Конфигурация сервера](#конфигурация-сервера)
5. [Запуск в Docker](#запуск-в-docker)
6. [Запуск на bare metal](#запуск-на-bare-metal)
7. [Мониторинг и метрики](#мониторинг-и-метрики)
8. [Масштабирование](#масштабирование)
9. [Резервное копирование и восстановление](#резервное-копирование-и-восстановление)
10. [Обновление](#обновление)
11. [Устранение неполадок](#устранение-неполадок)

## Архитектура системы

SSH Arena состоит из следующих компонентов:

- **Игровой движок (GameEngine)** — ядро, обрабатывающее игровую логику, тики, команды игроков.
- **SSH‑сервер** — принимает подключения игроков, аутентифицирует их, передаёт команды движку и отправляет обратно ASCII‑кадры.
- **Хранилище (SQLite)** — сохраняет состояние игры, прогресс игроков, мир.
- **Шина событий (EventBus)** — обеспечивает коммуникацию между компонентами (постройка здания, исследование, бой).
- **Рендерер (ANSIRenderer)** — генерирует ASCII‑представление мира для каждого игрока.
- **HTTP‑сервер метрик** — предоставляет Prometheus‑метрики и health‑check.

Все компоненты работают в одном процессе (монолит), что упрощает развёртывание.

## Требования к инфраструктуре

### Минимальные
- **CPU**: 1 ядро (x86‑64 или ARM64)
- **Память**: 512 МБ RAM
- **Диск**: 1 ГБ (для базы данных и логов)
- **ОС**: Linux (ядро 4.4+), macOS, Windows (с WSL2)
- **Сеть**: публичный IP‑адрес, открытый порт 2222 (SSH) и опционально 8080 (метрики)

### Рекомендуемые для 50+ игроков
- **CPU**: 2–4 ядра
- **Память**: 2–4 ГБ RAM
- **Диск**: 10 ГБ SSD
- **Пропускная способность**: 10 Мбит/с

## Настройка SSH‑аутентификации

### Генерация SSH‑ключей сервера

При первом запуске сервер автоматически создаёт ключ Ed25519 и сохраняет его в `./ssh_host_key`. Если вы хотите использовать свой ключ:

```bash
ssh-keygen -t ed25519 -f ssh_host_key -N ""
```

Убедитесь, что файл ключа недоступен для чтения посторонним:

```bash
chmod 600 ssh_host_key
```

### Авторизация игроков

Игроки подключаются по SSH с использованием публичных ключей. Сервер проверяет их по файлу `authorized_keys` (формат OpenSSH, одна строка — один ключ).

Пример:

```bash
# Сгенерируйте ключ на клиенте
ssh-keygen -t ed25519 -C "player1"

# Скопируйте публичный ключ в authorized_keys
echo "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ..." >> authorized_keys
```

Можно использовать несколько файлов, указав в конфиге:

```yaml
ssh:
  authorized_keys: "/etc/ssh/arena/authorized_keys"
```

### ForceCommand

Чтобы запретить игрокам выполнять произвольные shell‑команды, SSH Arena использует внутренний механизм ForceCommand (реализован в `internal/network/ssh_server.go`). При подключении игрок автоматически попадает в игровую сессию, и любые попытки выполнить команду вне игры блокируются.

## Конфигурация сервера

Основной файл конфигурации — `configs/config.yaml`. Все параметры описаны в README.md. Для production рекомендуется:

1. **Изменить порты** на стандартные, если нужно (например, SSH на 22).
2. **Настроить лимиты** (`max_players`, `tick_rate`).
3. **Включить подробное логирование** для отладки.
4. **Указать абсолютные пути** к ключам и базе данных.

Пример production‑конфига:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  max_players: 200
  tick_rate: 30
  save_interval: 60

game:
  world_width: 2000
  world_height: 2000
  chunk_size: 32
  starting_credits: 500
  tax_rate: 0.07

database:
  driver: "sqlite3"
  dsn: "/var/lib/ssh-arena/game.db"
  migrate: true
  pool_size: 20

logging:
  level: "info"
  format: "json"
  output: "/var/log/ssh-arena/server.log"
  with_caller: true

ssh:
  port: 2222
  private_key_path: "/etc/ssh/arena/ssh_host_key"
  authorized_keys: "/etc/ssh/arena/authorized_keys"
  banner: "Добро пожаловать в SSH Arena!"
  idle_timeout: 600
```

## Запуск в Docker

### Официальный образ

Официальный образ пока не опубликован. Вы можете собрать его самостоятельно:

```bash
docker build -t ssh-arena:latest .
```

### Docker Compose для production

Создайте `docker-compose.prod.yml`:

```yaml
version: '3.8'
services:
  ssh-arena:
    image: ssh-arena:latest
    container_name: ssh-arena
    restart: unless-stopped
    ports:
      - "2222:2222"
      - "8080:8080"
    volumes:
      - /etc/ssh/arena:/etc/ssh/arena:ro
      - /var/lib/ssh-arena:/var/lib/ssh-arena
      - /var/log/ssh-arena:/var/log/ssh-arena
    environment:
      - TZ=Europe/Moscow
    command: ["./ssh-arena", "--config", "/etc/ssh/arena/config.yaml"]
```

Запустите:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Оркестрация в Kubernetes

Пример манифеста Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ssh-arena
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ssh-arena
  template:
    metadata:
      labels:
        app: ssh-arena
    spec:
      containers:
      - name: ssh-arena
        image: ssh-arena:latest
        ports:
        - containerPort: 2222
          name: ssh
        - containerPort: 8080
          name: metrics
        volumeMounts:
        - mountPath: /var/lib/ssh-arena
          name: data
        - mountPath: /etc/ssh/arena
          name: config
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1"
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: ssh-arena-data
      - name: config
        configMap:
          name: ssh-arena-config
```

## Запуск на bare metal

### Установка как systemd‑сервиса

1. Соберите бинарник:

```bash
go build -o /usr/local/bin/ssh-arena ./cmd/server
```

2. Создайте пользователя и директории:

```bash
sudo useradd -r -s /bin/false ssh-arena
sudo mkdir -p /etc/ssh/arena /var/lib/ssh-arena /var/log/ssh-arena
sudo chown -R ssh-arena:ssh-arena /etc/ssh/arena /var/lib/ssh-arena /var/log/ssh-arena
```

3. Разместите конфиг и ключи в `/etc/ssh/arena`.

4. Создайте unit‑файл `/etc/systemd/system/ssh-arena.service`:

```ini
[Unit]
Description=SSH Arena game server
After=network.target

[Service]
Type=simple
User=ssh-arena
Group=ssh-arena
WorkingDirectory=/var/lib/ssh-arena
ExecStart=/usr/local/bin/ssh-arena --config /etc/ssh/arena/config.yaml
Restart=on-failure
RestartSec=5
StandardOutput=append:/var/log/ssh-arena/server.log
StandardError=append:/var/log/ssh-arena/error.log

[Install]
WantedBy=multi-user.target
```

5. Запустите:

```bash
sudo systemctl daemon-reload
sudo systemctl enable ssh-arena
sudo systemctl start ssh-arena
```

### Проверка статуса

```bash
sudo systemctl status ssh-arena
sudo journalctl -u ssh-arena -f
```

## Мониторинг и метрики

### Prometheus‑метрики

Сервер предоставляет метрики на эндпоинте `http://localhost:8080/metrics`. Доступные метрики:

- `ssh_arena_players_total` — текущее количество подключённых игроков.
- `ssh_arena_ticks_total` — общее число игровых тиков.
- `ssh_arena_commands_total` — количество обработанных команд.
- `ssh_arena_buildings_total` — общее число зданий в мире.
- `ssh_arena_resources_total` — количество каждого ресурса в экономике.
- `ssh_arena_uptime_seconds` — время работы сервера.

### Health‑check

Эндпоинт `http://localhost:8080/health` возвращает 200 OK, если сервер работает, и 503 в противном случае.

### Логи

Логи пишутся в JSON‑формате, что удобно для парсинга в ELK‑стеке. Пример записи:

```json
{
  "level": "info",
  "time": "2026-03-22T07:35:03Z",
  "caller": "engine/game.go:123",
  "msg": "Game engine started",
  "version": "0.1.0"
}
```

### Grafana‑дашборд

Пример дашборда для визуализации метрик можно найти в `scripts/grafana/dashboard.json`.

## Масштабирование

### Вертикальное масштабирование

Увеличьте параметры сервера (CPU, RAM) и настройки в конфиге:

- `max_players` — максимальное число игроков.
- `tick_rate` — частота тиков (чем меньше, тем выше нагрузка).
- `pool_size` в database — размер пула соединений с SQLite.

### Горизонтальное масштабирование (несколько комнат)

Текущая архитектура не поддерживает горизонтальное масштабирование «из коробки», но вы можете запустить несколько независимых экземпляров сервера на разных портах, используя общую базу данных (не рекомендуется из‑за блокировок SQLite). В будущем планируется поддержка шардинга по мирам.

### Балансировка нагрузки

Для распределения игроков между несколькими экземплярами можно использовать TCP‑балансировку (HAProxy, nginx) на порт SSH (2222). Каждый экземпляр должен иметь уникальный `server_id` в конфиге, чтобы избежать конфликтов сохранений.

## Резервное копирование и восстановление

### Что备份ировать

1. **База данных** — `/var/lib/ssh-arena/game.db` (основное состояние).
2. **Конфигурация** — `/etc/ssh/arena/config.yaml`.
3. **SSH‑ключи** — `/etc/ssh/arena/ssh_host_key` и `authorized_keys`.
4. **Логи** — `/var/log/ssh-arena/`.

### Автоматическое резервное копирование

Пример скрипта для ежедневного бэкапа (положить в `/etc/cron.daily/backup-ssh-arena`):

```bash
#!/bin/bash
BACKUP_DIR="/backup/ssh-arena"
DATE=$(date +%Y%m%d)
mkdir -p "$BACKUP_DIR/$DATE"
cp /var/lib/ssh-arena/game.db "$BACKUP_DIR/$DATE/game.db"
cp -r /etc/ssh/arena "$BACKUP_DIR/$DATE/config"
tar -czf "$BACKUP_DIR/$DATE.tar.gz" -C "$BACKUP_DIR/$DATE" .
rm -rf "$BACKUP_DIR/$DATE"
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete
```

### Восстановление из бэкапа

1. Остановите сервер.
2. Скопируйте `game.db` и конфиги на место.
3. Запустите сервер.

Если версия схемы базы данных изменилась, сервер автоматически выполнит миграции.

## Обновление

### Процедура обновления

1. **Сделайте бэкап** базы данных и конфигов.
2. **Остановите сервер** (systemctl stop ssh-arena или docker stop).
3. **Обновите код** (git pull) или скачайте новый релиз.
4. **Пересоберите** бинарник (go build) или образ Docker.
5. **Запустите сервер** — миграции применятся автоматически.
6. **Проверьте логи** на наличие ошибок.

### Откат

Если новая версия работает некорректно, восстановите старый бинарник и базу данных из бэкапа.

## Устранение неполадок

### Сервер не запускается

- **Проверьте конфиг** на синтаксические ошибки: `yamllint configs/config.yaml`.
- **Убедитесь, что порты не заняты**: `sudo lsof -i :2222`.
- **Проверьте права** на ключи и базу данных.

### Игроки не могут подключиться

- **Firewall**: `sudo ufw allow 2222/tcp`.
- **SSH‑ключи**: убедитесь, что публичный ключ игрока есть в `authorized_keys`.
- **Логи сервера** могут содержать детали ошибки аутентификации.

### Высокая загрузка CPU

- Увеличьте `tick_rate` (например, до 50 мс).
- Уменьшите частоту рендеринга (в коде `internal/ui/renderer.go`).
- Проверьте, не зациклились ли какие‑то игровые процессы (например, бесконечный цикл в экономике).

### Утечки памяти

Сервер написан на Go с автоматическим управлением памятью, но при длительной работе возможны утечки в кешах. Перезапускайте сервер раз в неделю или настройте автоматический restart в systemd.

### База данных заблокирована

SQLite не предназначен для высокой конкурентности. Если возникает ошибка `database is locked`:

- Увеличьте `pool_size`.
- Убедитесь, что нет других процессов, пишущих в тот же файл.
- Рассмотрите переход на PostgreSQL (в будущих версиях).

### Нет метрик на /metrics

Убедитесь, что в конфиге `server.port` указан корректный порт (по умолчанию 8080) и сервер слушает на всех интерфейсах (`0.0.0.0`).

## Контакты и поддержка

- **Issues**: [GitHub Issues](https://github.com/your-org/ssh-arena/issues)
- **Discord**: [ссылка на канал сообщества]
- **Документация**: [https://ssh-arena.readthedocs.io](опционально)

---
*Последнее обновление: 2026‑03‑22*