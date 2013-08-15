"""One-line documentation for timepusher module.

A detailed description of timepusher.
"""

import argparse
import Queue
import time
import threading
import sys
import urllib2
import fileinput
import json

namespace='test.namespace'
nssecret='5fbbd6p84o'
min_wait_in_sec=1
server='http://localhost:8080'
cookie_file='/home/leppoc/Downloads/timeengine (2).cookie'

#### End flags

push_url = server + '/api/timeseries/put'
cookie_file_content = open(cookie_file).read()

def send(obj):
  opener = urllib2.build_opener()
  opener.addheaders.append(('Cookie', cookie_file_content))
  opener.addheaders.append(('User-agent', 'timepusher'))
  r = opener.open(push_url, json.dumps(obj))

def make_data(lines):
  data={
      'ns': namespace,
      'nssecret': nssecret,
  }
  pts = []
  last_pt = None

  for l in lines:
    metric = l[0]
    value = float(l[1])
    date = int(l[2])
    resolution = len(l) == 4 and int(l[4]) or 1

    val = {
        'm': metric,
        'v': value,
    }

    # If same timestamp and resolution as previous point, reuse it.
    if last_pt and last_pt['t'] == date and last_pt['r'] == resolution:
      last_pt['vs'].append(val)
    else:
      # Otherwise, we make a new one.
      if last_pt:
        pts.append(last_pt)
      last_pt = {
          'r': resolution,
          't': date,
          'vs': [val],
      }
  if last_pt:
    pts.append(last_pt)

  data['pts'] = pts
  return data


def pusher():
  while True:
    start_time = time.clock()
    # Empty the queue.
    lines = []
    # break when either:
    # - (queue is empty AND we have at least one line)
    # - we have already 200 lines.
    while not (queue.empty() and len(lines) > 0 or len(lines) >= 200 or
               (queue.empty() and stop_pusher.is_set())):
      try:
        l = queue.get(True, 1)
        parts = l.split()
        if len(parts) == 3 or len(parts) == 4:
          lines.append(parts)
        else:
          print 'bad input:', l
      except Queue.Empty:
        pass

    # We may have no lines when the command is stopped.
    if len(lines) > 0:
      # Prepare data
      print "sending", len(lines), "lines"
      data = make_data(lines)
      # Send to backend
      send(data)

    # Check if we should still run.
    if stop_pusher.is_set() and queue.empty():
      print "bye thread"
      return

    end_time = time.clock()
    to_sleep = min_wait_in_sec - (end_time - start_time)
    time.sleep(to_sleep)


stop_pusher = threading.Event()
queue = Queue.Queue()

t = threading.Thread(target=pusher)
t.start()

while True:
  line = sys.stdin.readline()
  if line == '':
    break
  queue.put(line)

# Stop the pusher thread.
print "bye"
stop_pusher.set()
