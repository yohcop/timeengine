"""
Most useful doc:
https://developers.google.com/api-client-library/python/guide/aaa_oauth
"""

import argparse
import json
import Queue
import SocketServer
import sys
import threading
import time
import httplib2

from oauth2client.file import Storage
from oauth2client.client import flow_from_clientsecrets, SignedJwtAssertionCredentials
from apiclient.discovery import build

parser = argparse.ArgumentParser()
parser.add_argument('--namespace', '--ns',
                    required=True, help='namespace')
parser.add_argument('--secret',
                    required=True, help="Namespace's secret")
parser.add_argument('--port', default=0, type=int,
                    help="If non 0, starts a server on that port "
                    "listening for input. Otherwise, reads from stdin.")
parser.add_argument('--max_qps', default=1, type=float,
                    help="Maximum number of request per second to "
                         "send to the server.")
parser.add_argument('--server', default='http://localhost:8080',
                    help='URL for the server. Starts with http[s]://')
parser.add_argument('--max_push_size', default=200, type=int,
                    help='Maximum number of datapoints to send to '
                    'the server in a single request.')
parser.add_argument('--service_account',
                    help='Service account credentials')
parser.add_argument('--service_account_key',
                    help='Service account key')
parser.add_argument('--client_secret',
                    help='When not using a service account, path to the client '
                    'secret json file.')
args = parser.parse_args(sys.argv[1:])


queue = Queue.Queue()
stop_pusher = threading.Event()


class EndpointApi(object):
  def __init__(self, http):
    self.http = http
    self.service = build("timeengine", "v1",
        discoveryServiceUrl=("https://loonoscope.appspot.com/_ah/api/discovery/v1/"
                             "apis/{api}/{apiVersion}/rest"),
        http=http)

  def auth_service_account(self):
    credentials = SignedJwtAssertionCredentials(
            service_account_name=args.service_account,
            private_key=open(args.service_account_key).read(),
            scope="https://www.googleapis.com/auth/userinfo.email")
    credentials.authorize(self.http)

  def auth_user(self):
    storage = Storage('/tmp/cred')

    credentials = storage.get()

    if not credentials:
      flow = flow_from_clientsecrets(
        filename=args.client_secret,
        scope='https://www.googleapis.com/auth/userinfo.email',
        redirect_uri='urn:ietf:wg:oauth:2.0:oob')

      auth_uri = flow.step1_get_authorize_url()

      print 'Please go to:'
      print auth_uri
      code = raw_input("Type in the code you got after authorizing the app: ")

      credentials = flow.step2_exchange(code)
      storage.put(credentials)

    credentials.authorize(self.http)

  def send(self, obj):
    print obj
    try:
      result = self.service.put(body=obj).execute()
      print result
    except Exception as e:
      print sys.exc_info()[0]
      print e


def make_data(lines):
  data={
      'Ns': args.namespace,
      'NsSecret': args.secret,
  }
  pts = []
  last_pt = None

  for l in lines:
    metric = l[0]
    value = float(l[1])
    date = float(l[2])

    val = {
        'M': metric,
        'V': value,
        'T': date,
    }
    pts.append(val)

  data['Pts'] = pts
  return data


def pusher(api):
  def _pusher():
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
        print "sending", len(lines), "lines (%d)" % queue.qsize()
        data = make_data(lines)
        # Send to backend
        api.send(data)

      # Check if we should still run.
      if stop_pusher.is_set() and queue.empty():
        print "bye thread"
        return

      end_time = time.clock()
      to_sleep = (1.0/args.max_qps) - (end_time - start_time)
      if to_sleep > 0:
        time.sleep(to_sleep)
  return _pusher


def read_from_stdin():
  while True:
    line = sys.stdin.readline()
    if line == 'quitquitquit\n':
      break
    queue.put(line)


class SocketHandler(SocketServer.BaseRequestHandler):
  def handle(self):
    f = self.request.makefile()
    while True:
      line =  f.readline()
      if line == 'quitquitquit\n':
        print 'Received quitquitquit'
        self.server.shutdown()
        break
      if not line:
        break
      queue.put(line)


class ThreadedSocketHandler(SocketServer.ThreadingMixIn,
                            SocketServer.TCPServer):
  pass


def read_from_socket():
  server = ThreadedSocketHandler(('', args.port), SocketHandler)
  print "Listening on port", args.port
  try:
    server.serve_forever()
  except Exception, e:
    print "ERROR:", e


def main():
  http = httplib2.Http()
  api = EndpointApi(http)
  if args.service_account:
    api.auth_service_account()
  else:
    api.auth_user()

  t = threading.Thread(target=pusher(api))
  t.start()

  if args.port:
    t2 = threading.Thread(target=read_from_socket)
  else:
    t2 = threading.Thread(target=read_from_stdin)
  t2.start()
  try:
    t2.join()
  except:
    pass

  # Stop the pusher thread.
  print "bye"
  stop_pusher.set()

if __name__ == '__main__':
  main()
