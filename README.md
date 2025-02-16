## Тестовое задание: внутренний магазин мерча

Запуск в контейнере:

```
docker compose up
```

Локальный запуск:

```
make run
```

Запуск всех тестов:

```
make test
```

Просмотр тестового покрытия:

```
make test-cov
```

Запуск только интеграционных тестов:

```
make test-integration
```

## Нагрузочное тестирование

Проведено нагрузочное тестирование 1000 RPS в течение 200 секунд. Сценарий в `tests/load/main.go`.

### Результаты:

![результаты](tests/load/result.png)  
Доступность: 100%  
Среднее время ответа: 6 мс  
99% ответов: < 17 мс  
Максимальный ответ: 317 мс.

## Технологии:

-   Uber FX
-   HTTP Fiber
-   Postgres

## Вопросы:

1. Не было требований по уровню защиты от двойной траты. Выбрал оптимальное решение в плане цена/качество:
    - Repetable Read транзакции (полная защита от перезаписи)
    - Лог операций, благодаря которому после каждого изменения можно надежно пересчитать баланс
    - Запись пользователя в таблице с итоговым балансом также пересчитывается после изменения, но поскольку там на каждого пользователя уникальная строка, при попытке прочитать строчку, которая была модифирована другой транзакцией будет конфликт и откат второй транзакции, что в свою очередь исключает любой рассинхрон и двойные траты
2. Не понял формулировку "Сгруппированную информацию о перемещении монеток в его кошельке", при том, что в DTO поле называется coinHistory. В итоге решил сделал вывод агреггированной суммы по каждому контр-агенту. Если всё таки нужна именно история - модифировать можно буквально изменив одну строчку.
