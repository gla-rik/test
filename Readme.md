## DI

Install Wire by running: (для генерации файла зависимостей)
```sh
go install github.com/google/wire/cmd/wire@latest
```
and ensuring that `$GOPATH/bin` is added to your `$PATH`.

~~~sh 
cd .\internal\dependecy\
~~~

~~~sh 
wire
~~~

## 📊 API Endpoints

- `GET /ping` - проверка работоспособности
- `GET /api/orders` - список всех заказов
- `POST /api/orders` - создание нового заказа
- `GET /api/orders/:id` - получение заказа по ID
- `GET /api/orders/uid/:uid` - получение заказа по OrderUID
- `PUT /api/orders/:id` - обновление заказа
- `DELETE /api/orders/:id` - удаление заказа