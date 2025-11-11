# API Документация: Аутентификация

Документация для фронтенд-разработчиков мини-приложения МАКС.

## Общая информация

Все запросы должны отправляться с заголовком:
```
Content-Type: application/json
```

Для защищенных эндпоинтов требуется заголовок:
```
Authorization: Bearer <access_token>
```

---

## Эндпоинты аутентификации

### 1. POST `/auth/login`

**Описание:** Авторизация пользователя через данные Telegram Mini App. При успешной авторизации возвращает access token и информацию о пользователе.

**Требует авторизации:** Нет

**Что ожидает бэкенд:**

Request Body (JSON):
```json
{
  "query_id": "unique_session_id",
  "auth_date": 1633038072,
  "hash": "abc123def456...",
  "start_param": "start_parameter",
  "user": {
    "id": 123456789,
    "first_name": "Иван",
    "last_name": "Иванов",
    "username": "ivanov",
    "language_code": "ru",
    "photo_url": "https://example.com/photo.jpg"
  },
  "chat": {
    "id": -1001234567890,
    "type": "group"
  }
}
```

**Поля запроса:**
- `query_id` (string, обязательное) - уникальный идентификатор сессии
- `auth_date` (integer, обязательное) - Unix timestamp даты авторизации
- `hash` (string, обязательное) - хеш для проверки подлинности данных
- `start_param` (string, опциональное) - параметр запуска приложения
- `user` (object, обязательное) - объект пользователя Telegram
  - `id` (integer) - ID пользователя
  - `first_name` (string) - имя
  - `last_name` (string, опциональное) - фамилия
  - `username` (string, опциональное) - username
  - `language_code` (string, опциональное) - код языка
  - `photo_url` (string, опциональное) - URL фото
- `chat` (object, опциональное) - информация о чате (если приложение открыто в группе)
  - `id` (integer) - ID чата
  - `type` (string) - тип чата

**Что отдает бэкенд:**

Response 200 OK:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 123456789,
    "first_name": "Иван",
    "last_name": "Иванов",
    "username": "ivanov",
    "language_code": "ru",
    "photo_url": "https://example.com/photo.jpg"
  },
  "user_roles": ["student", "admin"]
}
```

**Поля ответа:**
- `access_token` (string) - JWT токен для доступа к защищенным эндпоинтам
- `user` (object) - информация о пользователе
- `user_roles` (array of strings) - роли пользователя в системе

**Cookies:**
Бэкенд автоматически устанавливает HTTP-only cookie:
- `refresh_token` - токен для обновления access token
  - Имя: `refresh_token`
  - HttpOnly: `true`
  - Secure: `false` (в development)
  - SameSite: `Lax`
  - Path: `/`

**Ошибки:**

Response 400 Bad Request:
```json
{
  "message": "Invalid request format"
}
```

Response 401 Unauthorized:
```json
{
  "message": "Invalid init data"
}
```
или
```json
{
  "message": "User not found"
}
```

---

### 2. POST `/auth/refresh`

**Описание:** Обновление access token с помощью refresh token. Используется для продления сессии пользователя.

**Требует авторизации:** Нет

**Что ожидает бэкенд:**

Request Body: Не требуется

**Cookies (входящие):**
Бэкенд ожидает HTTP-only cookie:
- `refresh_token` - токен для обновления access token (должен быть установлен ранее через `/auth/login`)

**Что отдает бэкенд:**

Response 200 OK:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Поля ответа:**
- `access_token` (string) - новый JWT токен для доступа к защищенным эндпоинтам

**Cookies (исходящие):**
Бэкенд автоматически обновляет HTTP-only cookie:
- `refresh_token` - новый refresh token (старый удаляется из базы данных)
  - Имя: `refresh_token`
  - HttpOnly: `true`
  - Secure: `false` (в development)
  - SameSite: `Lax`
  - Path: `/`

**Важно:**
- Старый refresh token удаляется из базы данных после использования (one-time use)
- Если refresh token истек или недействителен, необходимо заново авторизоваться через `/auth/login`

**Ошибки:**

Response 401 Unauthorized:
```json
{
  "message": "Invalid refresh token"
}
```
или
```json
{
  "message": "Expired refresh token"
}
```

Response 500 Internal Server Error:
```json
{
  "message": "Refresh cookie error"
}
```
или
```json
{
  "message": "System error"
}
```

---

### 3. GET `/auth/checkToken`

**Описание:** Проверка валидности текущего access token и refresh token. Возвращает статус токенов и информацию о пользователе.

**Требует авторизации:** Да (Bearer token)

**Что ожидает бэкенд:**

Headers:
```
Authorization: Bearer <access_token>
```

**Cookies (входящие, опционально):**
- `refresh_token` - для проверки валидности refresh token

**Что отдает бэкенд:**

Response 200 OK:
```json
{
  "access_token": {
    "valid": true
  },
  "refresh_token": {
    "valid": true
  },
  "user": {
    "username": "ivanov",
    "first_name": "Иван",
    "user_id": 123456789
  }
}
```

**Поля ответа:**
- `access_token.valid` (boolean) - валидность access token (всегда `true` при успешном запросе, так как запрос прошел через middleware)
- `refresh_token.valid` (boolean) - валидность refresh token из cookie
  - `true` - если refresh token существует, не истек и принадлежит текущему пользователю
  - `false` - если refresh token отсутствует, истек или не принадлежит пользователю
- `user.username` (string) - username пользователя (может быть пустой строкой, если username не установлен)
- `user.first_name` (string) - имя пользователя
- `user.user_id` (integer) - ID пользователя

**Ошибки:**

Response 401 Unauthorized:
```json
{
  "message": "Invalid or expired access token"
}
```
или
```
Missing authorization header
```
или
```
Invalid authorization format
```

---

## Работа с Cookies

### Установка cookies

Бэкенд автоматически устанавливает cookies в следующих случаях:

1. **POST `/auth/login`** - устанавливает `refresh_token` cookie
2. **POST `/auth/refresh`** - обновляет `refresh_token` cookie

### Параметры cookies

Все cookies устанавливаются со следующими параметрами:
- **Имя:** `refresh_token`
- **HttpOnly:** `true` (недоступен из JavaScript)
- **Secure:** `false` (в development окружении)
- **SameSite:** `Lax`
- **Path:** `/`

### Отправка cookies

Для работы с cookies необходимо:
- Отправлять все запросы с параметром `credentials: 'include'` (или эквивалентом в используемом HTTP клиенте)
- Это позволяет браузеру автоматически отправлять cookies при каждом запросе

### Особенности в мини-приложении МАКС

В мини-приложениях Telegram cookies могут работать некорректно в некоторых случаях. Убедитесь, что:
- Все запросы отправляются с поддержкой cookies
- При работе в iframe могут быть ограничения на работу с cookies
- Если cookies не работают, может потребоваться альтернативный способ хранения refresh token

---

## Обработка ошибок

Все ошибки возвращаются в формате:
```json
{
  "message": "Описание ошибки",
  "error": "Детальная информация об ошибке (опционально)"
}
```

**Коды статусов:**
- `200` - Успешный запрос
- `400` - Неверный формат запроса
- `401` - Неавторизован (неверный/истекший токен, отсутствует заголовок Authorization)
- `500` - Внутренняя ошибка сервера

---

## Схема работы токенов

1. **Access Token:**
   - Выдается при `/auth/login` и `/auth/refresh`
   - Передается в заголовке `Authorization: Bearer <token>`
   - Имеет ограниченный срок действия (обычно несколько часов)
   - Используется для доступа к защищенным эндпоинтам
   - Хранится на клиенте (localStorage, sessionStorage и т.д.)

2. **Refresh Token:**
   - Выдается при `/auth/login` и обновляется при `/auth/refresh`
   - Хранится в HTTP-only cookie `refresh_token`
   - Имеет более длительный срок действия (обычно несколько дней)
   - Используется только для обновления access token
   - Недоступен из JavaScript (HttpOnly)
   - Удаляется из базы данных после использования (one-time use)

3. **Проверка токенов:**
   - `/auth/checkToken` проверяет валидность обоих токенов
   - Access token проверяется через JWT middleware
   - Refresh token проверяется через базу данных

---

## Рекомендации

1. **Хранение Access Token:**
   - Сохраняйте access token на клиенте (localStorage, sessionStorage, memory)
   - Не передавайте access token в URL параметрах
   - Используйте HTTPS для передачи токенов

2. **Обновление токенов:**
   - Реализуйте автоматическое обновление access token перед истечением срока действия
   - При получении 401 ошибки попробуйте обновить токен через `/auth/refresh`
   - Если refresh token истек, перенаправьте пользователя на повторную авторизацию

3. **Проверка токенов:**
   - Используйте `/auth/checkToken` при инициализации приложения для проверки состояния сессии
   - Проверяйте валидность refresh token перед попыткой обновления access token

4. **Безопасность:**
   - Не логируйте токены
   - Не передавайте токены в URL
   - Используйте HTTPS в production
   - Реализуйте защиту от CSRF атак
