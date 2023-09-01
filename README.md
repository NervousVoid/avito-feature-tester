# Тестовое задание для стажёра Backend
# Сервис динамического сегментирования пользователей

##### Автор: [Андросов Петр](https://t.me/nervous_void) 

[Здесь](problem.md) можно найти полный текст задания

### Запуск
```shell
  docker-compose up
```
#### Запуск, если в прошлый раз что-то пошло не по плану
```shell
  docker rm $(docker ps -a -q) && docker volume prune -f
  docker rmi -f avito-segmentator
  docker-compose up
```
После успешного запуска контейнеров, в базе данных будут созданы 1000 пользователей, а таблицы сегментов и связи сегментов с пользователями будут пустыми

### Доступные методы

*У проекта есть [Swagger-файл](docs/swagger.yaml) и описание методов в [Postman](https://red-water-385938.postman.co/workspace/Peter-Androsov-Workspace~74fa4139-afcf-49bf-8b7f-4a31ffdb000b/collection/8903220-80f256d1-e22d-476b-8312-89794e8caf97?action=share&creator=8903220)*

#### **POST** /api/create_segment
Метод создания нового сегмента

Принимает *опцианальный* параметр **Fraction**, при задании которого, создаваемый сегмент будет автоматически присваиваться случйному заданному проценту пользователей

*В примере ниже сегмент 30-процентной скидки будет создан и автоматически присвоен 10% пользователей*

*Принимаемая структура*
```json
{
  "segment_slug": "AVITO_DISCOUNT_30",
  "fraction": 10
}
```
или, если нужно просто создать сегмент:
```json
{
  "segment_slug": "AVITO_DISCOUNT_30"
}
```

#### **DELETE** /api/delete_segment
Метод удаления сегмента

*Принимаемая структура*
```json
{
  "segment_slug": "AVITO_VOICE_MESSAGES"
}
```
  
#### **POST** /api/update_user_segments
Метод обновления данных о сегментах у юзера\
Принимает id пользователя, сегменты, в которые нужно добавить пользователя, и из которых убрать

*Принимаемая структура*
```json
{
  "user_id": 1234,
  "assign_segments": ["AVITO_DISCOUNT_30"],
  "unassign_segments": [
    "AVITO_DISCOUNT_50",
    "AVITO_VOICE_MESSAGES"
  ]
}
```

#### **GET** /api/get_user_segments
Метод получения активных сегментов пользователя

*Принимаемая структура*
```json
{
  "user_id": 1002
}
```
*Возвращаемая структура*
```json
{
  "segments": ["AVITO_DISCOUNT_30","AVITO_DISCOUNT_50"],
  "user_id": 1002
}
```

#### **GET** /api/get_user_history
Метод получения активных сегментов пользователя
Принимает id пользователя, а также границы временного промежутка в форматах "YYYY-MM" или "YYYY-M"

Возвращает ссылку на отчет в формате .csv

*Принимаемая структура*
```json
{
  "user_id": 1000,
  "start_date": "2023-5",
  "end_date": "2023-9"
}
```
*Возвращаемая структура*
```json
{
  "csv_url": "0.0.0.0:8000/reports/report_k6cyy3f25a.csv"
}
```