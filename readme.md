### A simple worker for sending periodic heart beat signals to all machines via rabbit exchange.

#### Role
This is `alti-heartbeater`, which is responsible for steps 1, 3 and 4 below.

#### Listen to queues
* `altizure-heart-pong`

#### Publish to exchange (fanout)
* `altizure-heart-ping`

#### Flow
1. `alti-heartbeater` sends out ping message to fanout exchange `altizure-heart-ping`
2. Any subscriber should responds with a pong message and publish to `alti-heart-pong`
3. `alti-heartbeater` will get back the pong messages and determine if any machine goes up or down
4. Current status is stored in database: `hearbeat`; collection: `machine`

### To subscribe
Set the followings in .env:
```bash
RABBIT_EXCHANGE_ALTI_HEART_PING=altizure-heart-ping
RABBIT_QUEUE_ALTI_HEART_PONG=altizure-heart-pong
ALTI_HOST_TYPE=alti-transferer
ALTI_HOST_NAME=citymap
ALTI_NICK_NAME=Tangela
```
Then call:
```go
SubscribeHeartbeat(conn, ch, pingExchangeName, pongQueueName)
```

#### Sample ping
```bash
{"name":"","nickname":"","type":"","ping":"2016-12-15 07:08:08.0004 PM","pong":"0001-01-01 12:00:00.0000 AM","response":null,"extra":"","status":""}
```

#### Sample pong
```bash
{"name":"eserver3","nickname":"Eevee3","type":"alti-archiver","ping":"2016-12-15 07:08:08.0004 PM","pong":"2016-12-15 07:08:08.0491 PM","extra":""}
```

#### Log
```bash
sudo journalctl -o cat -f -u alti-heartbeater
```
