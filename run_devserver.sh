#!/bin/sh

~/slash/opt/google_appengine/dev_appserver.py --host=0.0.0.0 $* src/app.yaml
