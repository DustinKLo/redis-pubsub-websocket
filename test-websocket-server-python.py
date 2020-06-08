import os
import random
from threading import Thread

import asyncio
import websockets
import redis

CLIENTS = set()  # set of clients we keep track of

redis_conn = redis.Redis(host='localhost', port=6379, db=0)
pubsub = redis_conn.pubsub()
pubsub.subscribe("test")


async def redis_subscribe_task():
    # params: not sure, maybe itll use the global CLIENTS set(?)
    # will be a asyncio task that will send messages through the CLIENTS variable
    #         ex. subscribe_task = asyncio.create_task(redis_subscribe_task(...))
    for message in pubsub.listen():
        print(message)
        for client in CLIENTS:
            print("data", message['data'])
            await client.send(message['data'].decode("utf-8"))

def redis_wrapper_func(loop):
        # running blocking function, pubsub.listen() in separate thread
        asyncio.set_event_loop(loop)
        loop.run_until_complete(redis_subscribe_task())


async def hello(websocket, path):
    try:
        # got connection from websocket client
        print("client connected: ", websocket)
        CLIENTS.add(websocket)
        print("all clients: ", CLIENTS)

        async for message in websocket:  # reacting to messages from client
            print(message)
            for client in CLIENTS:
                print("sending message to client", client)
                await client.send(message)

    finally:
        print("unregistering client", websocket)
        CLIENTS.remove(websocket)



loop = asyncio.new_event_loop()
t = Thread(target=redis_wrapper_func, args=(loop,))
t.start()

PORT = os.getenv("PORT") or 8765
start_server = websockets.serve(hello, "localhost", PORT)

print("running server on port {}".format(PORT))
asyncio.get_event_loop().run_until_complete(start_server)
asyncio.get_event_loop().run_forever()
