
P=`pwd`

TP="$P/third_party"
#for f in $P/third_party/src/*; do
#  if [ "" == "$TP" ]; then
#    TP=$f
#  else
#    TP=$TP:$f
#  fi
#done

export GOROOT=/home/leppoc/slash/opt/google_appengine/goroot
export GOPATH=$TP:$P/genfiles:$P
