# Builds the pusher script in build/endpoint_pusher

pyinstaller \
  --clean \
  --onefile \
  --specpath=/tmp/endpoint_pusher/spec \
  --distpath=src/timeengine/static \
  --workpath=/tmp/endpoint_pusher/build \
  src/timeengine/static/endpoint_pusher.py
