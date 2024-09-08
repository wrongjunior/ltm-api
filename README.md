
# LTM API

## Описание
Проект **LTM API** представляет собой API для оценки времени чтения текста на основе метода Флеша-Кинкейда. Сервис позволяет загружать текстовые файлы или напрямую передавать текст в запросе, для получения оценки времени чтения, количества слов, предложений, слогов и индекса удобочитаемости.

## Демонстрационное видео
Видеопример работы проекта можно найти 

### Основные функции:
- Оценка времени чтения текста.
- Поддержка параллельной обработки текста для повышения производительности.
- Оценка текста как по переданному текстовому запросу, так и через загруженные файлы.

## Структура проекта
Проект состоит из нескольких основных пакетов:
- `config`: Управление конфигурацией приложения.
- `router`: Определение маршрутов и создание HTTP роутера.
- `estimator`: Логика оценки текста (время чтения, индекс удобочитаемости и т.д.).
- `api`: Обработчики HTTP запросов для оценки текста.

## Установка и запуск

1. Склонируйте репозиторий:

   ```bash
   git https://github.com/wrongjunior/ltm-api.git
   ```

2. Перейдите в папку проекта:

   ```bash
   cd ltm-api
   ```

3. Установите зависимости и выполните сборку проекта:

   ```bash
   go mod download
   go build
   ```

4. Запустите сервер:

   ```bash
   PORT=8080 ./ltm-api
   ```

   Сервер запустится на порту 8080 (либо на другом, если указан в переменной окружения `PORT`).

## Маршруты API

- `POST /estimate/reading-time`: Оценка времени чтения для переданного текста.
    - **Параметры**: JSON объект с полями:
        - `text`: Текст для оценки.
        - `readingSpeed`: Скорость чтения (слов в минуту).
        - `hasVisuals`: Учитывать ли наличие визуальных элементов (увеличивает время чтения).
        - `workerCount`: Количество воркеров для параллельной обработки.

- `POST /estimate/upload`: Оценка времени чтения для загруженного файла.
    - **Параметры**: Файл с текстом для оценки.

## Лицензия
Проект распространяется под лицензией MIT. Подробности можно найти в файле [LICENSE](./LICENSE).

