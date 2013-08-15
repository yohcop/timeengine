"""One-line documentation for walk module.

A detailed description of walk.
"""

import random
import time
import sys

pause=1
num_metrics=10
max_out=72*60*60
time_delta=-72*60*60
live=True

def walk(n):
  return n + random.random() - 0.5


# Main =============================
vals=[0] * num_metrics
time_start=int(time.time())-time_delta
while max_out > 0:
  for i, v in enumerate(vals):
    if live:
      d=int(time.time())
    else:
      d=time_start
    vals[i] = walk(v)
    print "my.metric.%d %f %d" % (i, vals[i], d)
    sys.stdout.flush()

  time_start+=1
  if max_out > 0:
    max_out -= 1
    time.sleep(pause)
