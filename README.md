<h4>Dockerize application</h4>

```
docker build -t redis-pubsub-websocket:latest .
docker run -p 8000:8000 redis-pubsub-websocket:latest
```

<h4>Infratructure diagram</h4>

![diagram](redis-pubsub-websocket.png)

<h3>Connecting to Websocket Server</h3>
```
// specify which "rooms" you want to subscribe to as comma separated URL params

// in your browser console
var ws = new Websocket("ws://localhost:8000/room1,room2,room3");
ws.onmessage = e => console.log(e.data);
ws.onclose = e => console.log(e);
```
<h5>Sending message to Redis PubSub</h5>
```
$ redis-cli
> PUBLISH room1 "hello world room1!!"
```
