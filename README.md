### Infratructure diagram

![diagram](./img/redis-pubsub-websocket.png)

### Running the websocket server

```bash
$ go run *.go

# or

$ go build .
$ ./redis-pubsub-websocket
```

#### Command line arguments

```bash
$ ./redis-pubsub-websocket --help
Usage of ./redis-pubsub-websocket:
  -debug
      debug mode, stdout results
  -redis string
      redis endpoint (default: redis://127.0.0.1:6379) (default "redis://127.0.0.1:6379")
```

### Connecting to the websocket server with Javascript

```javascript
// specify which "rooms" you want to subscribe to as comma separated URL params
// in your browser console
var ws = new Websocket("ws://localhost:8000/room1,room2,room3");
ws.onmessage = (e) => console.log(e.data);
ws.onclose = (e) => console.log(e);
```

### Testing Websocket server

- Go to `http://localhost:8000`

![homepage](./img/websocket-test.png)

###### Sending message to Redis PubSub

```bash
$ redis-cli
PUBLISH testroom "hello world room1!!"
```

### Using Docker

##### Building the image

```bash
$ docker build -t redis-pubsub-websocket:latest .
$ docker run -p 8000:8000 redis-pubsub-websocket:latest
```

##### Running container

- Use `-redis redis://host.docker.internal:6379` to connect to your local machine's redis server

```bash
$ docker run -p 8000:8000 redis_pubsub_websocket -redis redis://host.docker.internal:6379 -debug
```
