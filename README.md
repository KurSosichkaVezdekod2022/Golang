# Golang

## Запуск
* Задания 10-30 запускаются на test.txt в корне репозитория
    ```
    go run/build x.go
    ```

* Чтобы запустить сервер на 40 и тесты к нему, нужно  сначала привести в рабочее состояние сервер (40.go) и выполнить
    ```
    go test 40_test.go
    ```
* Задание 50 перебиравет все proto и pb файлы из директорий 50/proto и 50/pb соответственно
    ```
    go run 50.go
    ```

## Немного о решениях
Решения заданий 20-40 используют горутины, каналы и sync.Waitgroup для синхронизации.

В решении задания на 50 сначала используется построение структур сообщений с помощью парсинга proto файлов и борьбы с типами, а потом рекурсивное заполнение их данными из pb файлов с проверками на соответствие структуре. На данный момент некоторые структуры читаются в строки, что вряд ли задумывалось. Вывод имеет вид:
```
<proto файл> <pb файл>
<имя поля> : <значение>
<имя поля> : <значение>
...
```


