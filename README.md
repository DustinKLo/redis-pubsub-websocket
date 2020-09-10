### Infratructure diagram
![diagram](./img/redis-pubsub-websocket.png)

### Running the websocket server
###### Using Docker
```
$ docker build -t redis-pubsub-websocket:latest .
$ docker run -p 8000:8000 redis-pubsub-websocket:latest
```
###### Not using Docker
```
$ go run main.go redis.go hub.go client.go
```

### Connecting to the websocket server with Javascript
```
// specify which "rooms" you want to subscribe to as comma separated URL params
// in your browser console
var ws = new Websocket("ws://localhost:8000/room1,room2,room3");
ws.onmessage = e => console.log(e.data);
ws.onclose = e => console.log(e);
```

### Testing Websocket server
###### Go to `http://localhost:8000`
![homepage](./img/websocket-test.png)

###### Sending message to Redis PubSub
```
$ redis-cli
PUBLISH testroom "hello world room1!!"
```
