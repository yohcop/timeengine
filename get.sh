
data="{\"Serie\":{\"r\":1,\"t\":$2,\"to\":$3,\"m\":\"$4\"}}"

curl "$url" \
  -d "$data" \
  -H "Accept:text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" \
  -H "Accept-Encoding:gzip,deflate,sdch" \
  -H "Accept-Language:en-US,fr;q=0.8" \
  -H "Cache-Control:max-age=0" \
  -H "Connection:keep-alive" \
  -H "Cookie:dev_appserver_login="test@example.com:False:185804764220139124118"" \
  -H "Host:localhost:8080" \
  -H "User-Agent:Mozilla/5.0 (X11; Linux i686) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/29.0.1547.41 Safari/537.36"

