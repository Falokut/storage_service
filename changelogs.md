## v1.2.1
* Удалёна настройка `max_range_request_length`
* Обновлён go-kit
* Фикс обработки ошибок из minio
## v1.2.0
* Добавлено вычисление размера файла, раньше размер файла брался из `http.Request.ContentLength`
## v1.1.0
* Теперь MaxRequestBodySize берётся из параметра конфига `max_file_size`
* `max_file_size` стал обязательным параметром
* Удалена неиспользуемая переменная `base_local_storage_path`
## v1.0.0
* Инициализация проекта