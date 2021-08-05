# Пример вывода клиента
```
pow_client_1  | sending gimme request...
pow_client_1  | pow challenge: UVg42gMtUv-3-ae42df0422d745e55512c86f0c2e995e3c32cbc86e59036c5ba488b4c1f4b866
pow_client_1  | solution found: UVg42gMtUv-5089239-ae42df0422d745e55512c86f0c2e995e3c32cbc86e59036c5ba488b4c1f4b866
pow_client_1  | hash : 000�D]JKW���\Ev�m\�����R ����,�L�;�
pow_client_1  | time spent: 11.7048415s
pow_client_1  | ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
pow_client_1  | The world is round so that friendship may encircle it. (Pierre Teilhard de Chardin)
pow_client_1  | ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

# Как запустить

Первые два шага можно пропустить если нет [проблем с шагом 3](https://github.com/docker/buildx/issues/476)
1. docker pull alpine:3.14
2. docker pull golang:1.16
3. cd /build && docker-compose up

# Упрощения
- сервер и клиент написаны максимально просто
- нет распараллеливания при поиске решения, там это имеет смысл
- порты/хосты захардкожены, в реальном приложении конечно так делать нельзя
- тесты не претендуют на полноту покрытия кейсов, скорее написаны для меня самого, так проще проверить
правильность, чем делать это руками

# Как работает

1. Сервер при запросе `gimme` генерирует задачу - **найти такое x при котором hash(challenge_data + x)
содержит N ведущих нулей**
2. Сервер генерирует mac, чтобы можно было проверить не подделал ли клиент данные задачи
3. Сервер отправляет клиенту задачу в виде nonce-N-mac
4. Клиент решает задачу перебором от 0 до uint32_MAX, дописывая к nonce байтовое представление числа и высчитывая хеш
5. Когда решение найдено клиент отвечает серверу текстом `gimme nonce-solution-mac`
6. Сервер снова считает mac по переданному клиентом nonce и сравниает с переданным клиентом mac, если сходится - задача 
действительно была сгенерирована сервером
7. Сервер высчитывает хеш по nonce и переданному решению и проверяет условие задачи
8. Если все ок, отдает цитату в виде строки

Сервер старается как можно быстрее закрыть tcp соединение,  поэтому клиенту приходится каждый раз открывать новое.
Получить задачу - открываем соединение. Отправить решение - новое соединение. Это логично, иначе в чем смысл
pow, если злоумышленник может просто долбить запросами на получение задачи и не решать ее, но я решил все таки этот
момент тоже в ридми уточнить

# Есть одна проблема

На этапе дизайна я не подумал про защиту от использования уже решенных задач. Как итог, можно решить одну задачу и
бесконечно много использовать ее решение. Доделывать не буду, посколько уже пообещал отправить код на проверку.
Но как бы я это сделал: добавляется еще один параметрт - таймштамп момента генерации задачи. На сервере
выставялется параметр - ttl задачи. Таймштмап используется также в процессе генерации mac, чтобы защититься 
от изменения на стороне клиента, т.к. таймштмап передается клиенту, и клиент потом его возвращает вместе с другими данными
(чтобы не хранить состояние на сервере). 
После получения решения и валидации mac сервер отбрасывает сообщение если ts + ttl <= now.
Остается еще один вопрос: если решение получено быстро, то его можно использовать до момента протухания ttl.
Для решения этой проблемы на стороне сервера можно хранить список решенных задач, который очищается раз в ttl.
Для оптимизации используемой памяти можно использовать вероятностные структуры данных, например bloom-фильтр

# Бенчмаркинг

Просто чтобы убедиться что поиск решения - намного затратнее, чего его проверка.

**L = длина nonce, N - количество требуемых нулей**

## Поиск решения

```
N = 1     L = 5      BenchmarkServerSolve_1_5-12           78872             17311 ns/op
N = 2     L = 10     BenchmarkServerSolve_2_10-12            100          24911305 ns/op
N = 2     L = 20     BenchmarkServerSolve_2_20-12            160          21156350 ns/op
N = 2     L = 50     BenchmarkServerSolve_2_50-12            373          65324338 ns/op
```

## Проверка решения

```
N = 1     L = 5      BenchmarkServerValidate_1_5-12           1528162               791 ns/op
N = 2     L = 10     BenchmarkServerValidate_2_10-12          1554482               759 ns/op
N = 2     L = 20     BenchmarkServerValidate_2_20-12          1573590               762 ns/op
N = 2     L = 50     BenchmarkServerValidate_2_50-12          1321009               914 ns/op
```