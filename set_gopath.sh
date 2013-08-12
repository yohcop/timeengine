
P=.

TP=""
for f in $P/third_party/*; do
  if [ "" == "$TP" ]; then
    TP=$f
  else
    TP=$TP:$f
  fi
done

export GOROOT=/home/leppoc/slash/opt/google_appengine/goroot
export GOPATH=$TP:$P/genfiles:$P
