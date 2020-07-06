import json
import redis
import uuid
import datetime
import time
import random

# connect with redis server as Alice
r = redis.Redis(host='localhost', port=6379, db=0)

lim = 10000

start = time.time()

rooms = ['room1', 'room2', 'room3', 'room4', 'room5', 'room6']

for i in range(lim):
  room = random.choice(rooms)
  if i % 100 == 0:
    print(i)
  d = {
    # 'counter': i,
    'id': str(uuid.uuid4()),
    'timestamp': datetime.datetime.now().isoformat(),
    'room': room
  }
  time.sleep(random.uniform(0, 0.000000))
  r.publish(room, json.dumps(d))

print(time.time() - start, "seconds")
