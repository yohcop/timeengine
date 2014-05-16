"""One-line documentation for walk module.

A detailed description of walk.
"""

import argparse
import random
import time
import sys

parser = argparse.ArgumentParser()
parser.add_argument('--metrics', default=3, type=int,
                    help="Number of metrics per second to generate.")
parser.add_argument('--start', default=time.time(), type=float,
                    help="Start timestamp.")
parser.add_argument('--stop', default=time.time() + 24*60*60, type=float,
                    help="End timestamp.")
parser.add_argument('--step', default=1, type=float,
                    help="Time step, used when --live is false.")
parser.add_argument('--pause', default=1, type=float,
                    help="Pause in seconds between metric generation")
parser.add_argument('--live', default=1, type=int,
                    help="If true, then --start is ignored, and current time "
                    "is used instead")

args = parser.parse_args(sys.argv[1:])

pause=args.pause
num_metrics=args.metrics
time_start=args.start
time_end=args.stop
live=args.live

def walk(n):
  return n + random.random() - 0.5


# Main =============================
vals=[0] * num_metrics
while time_end > time_start:
  for i, v in enumerate(vals):
    if live:
      d=int(time.time())
    else:
      d=time_start
    print "my.metric.%d %f %f" % (i, vals[i], d)
    vals[i] = walk(v)

  sys.stdout.flush()

  if not live:
    time_start += args.step

  time.sleep(pause)
