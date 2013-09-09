"""One-line documentation for timepusher module.

A detailed description of timepusher.
"""

import argparse
import cookielib
import getpass
import json
import Queue
import sys
import threading
import time
import urllib
import urllib2

parser = argparse.ArgumentParser()
parser.add_argument('--namespace', '--ns',
                    required=True, help='namespace')
parser.add_argument('--secret',
                    required=True, help="Namespace's secret")
parser.add_argument('--max_qps', default=1, type=float,
                    help="Maximum number of request per second to "
                         "send to the server.")
parser.add_argument('--server', default='http://localhost:8080',
                    help='URL for the server. Starts with http[s]://')
parser.add_argument('--max_push_size', default=200, type=int,
                    help='Maximum number of datapoints to send to '
                    'the server in a single request.')
parser.add_argument('--cookie_jar', default='~/.config/timepusher',
                    help='Cookie jar path.')
parser.add_argument('--dev_cookie',
                    help='For dev only. Cookie content for auth.')
args = parser.parse_args(sys.argv[1:])

#### End flags

push_url = args.server + '/api/timeseries/put'
app_name = "timeengine"
verify_auth_url = args.server + '/checkauth'

cookiejar = cookielib.LWPCookieJar(args.cookie_jar)
opener = urllib2.build_opener(urllib2.HTTPCookieProcessor(cookiejar))
urllib2.install_opener(opener)

queue = Queue.Queue()
stop_pusher = threading.Event()


def auth():
  try:
    cookiejar.load()
  except:
    pass

  if isAuth():
    return True
  else:
    email = raw_input('Email: ')
    password = getpass.getpass()

  auth_uri = 'https://www.google.com/accounts/ClientLogin'
  authreq_data = urllib.urlencode({ "Email":   email,
                                    "Passwd":  password,
                                    "service": "ah",
                                    "source":  app_name,
                                    "accountType": "HOSTED_OR_GOOGLE" })
  auth_req = urllib2.Request(auth_uri, data=authreq_data)
  auth_resp = urllib2.urlopen(auth_req)
  auth_resp_body = auth_resp.read()
  auth_resp_dict = dict(x.split("=")
                        for x in auth_resp_body.split("\n") if x)
  authtoken = auth_resp_dict["Auth"]

  serv_args = {}
  serv_args['continue'] = verify_auth_url
  serv_args['auth']     = authtoken

  full_serv_uri = args.server + "/_ah/login?%s" % (
      urllib.urlencode(serv_args))

  serv_req = urllib2.Request(full_serv_uri)
  serv_resp = urllib2.urlopen(serv_req)
  serv_resp_body = serv_resp.read()

  cookiejar.save()
  print serv_resp_body
  return 'ok' == serv_resp_body

def isAuth():
  req = urllib2.Request(verify_auth_url)
  if args.dev_cookie:
    req.add_header('Cookie', args.dev_cookie)
  resp = urllib2.urlopen(req)
  serv_resp_body = resp.read()
  cookiejar.save()
  return 'ok' == serv_resp_body

def send(obj):
  d = json.dumps(obj)
  try:
    req = urllib2.Request(push_url, d)
    r = urllib2.urlopen(req)
    print r.getcode()
  except urllib2.URLError, e:
    print e

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
    while not (queue.empty() and len(lines) > 0
               or len(lines) >= args.max_push_size
               or (queue.empty() and stop_pusher.is_set())):
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
    if to_sleep < 0:
      to_sleep = 0
    time.sleep(to_sleep)


def main():
  if not auth():
    print "Could not authenticate."
    return

  t = threading.Thread(target=pusher)
  t.start()

  while True:
    line = sys.stdin.readline()
    if line == 'quitquitquit\n':
      break
    queue.put(line)

  # Stop the pusher thread.
  print "bye"
  stop_pusher.set()

main()
