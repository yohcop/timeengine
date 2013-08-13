#!/bin/bash
#set -x

server=$1
pause=$2
ns=$3
secret=$4
metric=$5

last=0

while [ 1 ]; do
  n=`expr \( $RANDOM % 101 \) - 50`
  last=`expr $last + $n`
  url="$server/api/put"
  data="{\"ns\":\"$ns\",\"nssecret\":\"$secret\",\"Pts\":[{\"r\":1,\"t\":`date +%s`,\"Vs\":[{\"m\":\"$metric\",\"v\":$last}]}]}"
  echo $data
  curl "$url" \
    -d "$data" \
    -H "Cookie:dev_appserver_login=\"test@example.com:False:185804764220139124118\""

  sleep $pause
done

#    -H "Accept:text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" \
#    -H "Accept-Encoding:gzip,deflate,sdch" \
#    -H "Accept-Language:en-US,fr;q=0.8" \
#    -H "Cache-Control:max-age=0" \
#    -H "Connection:keep-alive" \
#    -H "Host:localhost:8080" \
#    -H "User-Agent:Mozilla/5.0 (X11; Linux i686) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/29.0.1547.41 Safari/537.36"
