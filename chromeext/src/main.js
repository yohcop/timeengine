var tcpServer;
var commandWindow;

/**
 * Listens for the app launching then creates the window
 *
 * @see http://developer.chrome.com/trunk/apps/app.runtime.html
 * @see http://developer.chrome.com/trunk/apps/app.window.html
 */
chrome.app.runtime.onLaunched.addListener(function() {
        if (commandWindow && !commandWindow.contentWindow.closed) {
                commandWindow.focus();
        } else {
                chrome.app.window.create('index.html',
                        {id: "mainwin", bounds: {width: 500, height: 309, left: 0}},
                        function(w) {
                                commandWindow = w;
                        });
        }
});


// event logger
var log = (function(){
  var logLines = [];
  var logListener = null;

  var output=function(str) {
    if (str.length>0 && str.charAt(str.length-1)!='\n') {
      str+='\n'
    }
    logLines.push(str);
    if (logListener) {
      logListener(str);
    }
  };

  var addListener=function(listener) {
    logListener=listener;
    // let's call the new listener with all the old log lines
    for (var i=0; i<logLines.length; i++) {
      logListener(logLines[i]);
    }
  };

  return {output: output, addListener: addListener};
})();

var put = (function(){
    var frame = null;

    var send=function(raw) {
      if (frame) {
        frame.contentWindow.postMessage(raw, "*");
      }
    };

    var setFrame=function(frameObj) {
      frame = frameObj;
    };
    return {send:send, setFrame: setFrame};
})();

function onAcceptCallback(tcpConnection, socketInfo) {
  var info="["+socketInfo.peerAddress+":"+socketInfo.peerPort+"] Connection accepted!";
  log.output(info);
  console.log(socketInfo);
  tcpConnection.addDataReceivedListener(function(data) {
    try {
      put.send(data);
    } catch (ex) {
      log.output(ex);
    }

    var lines = data.split(/[\n\r]+/);
    var info="["+socketInfo.peerAddress+":"+socketInfo.peerPort+"] ";
    log.output(info + "got " + lines.length + " lines");
    //for (var i=0; i<lines.length; i++) {
    //  var line=lines[i];
    //  if (line.length>0) {
    //    var info="["+socketInfo.peerAddress+":"+socketInfo.peerPort+"] "+line;
    //    log.output(info);
    //  }
    //}
  });
};

function startServer(addr, port) {
  if (tcpServer) {
    tcpServer.disconnect();
  }
  tcpServer = new TcpServer(addr, port);
  tcpServer.listen(onAcceptCallback);
}


function stopServer() {
  if (tcpServer) {
    tcpServer.disconnect();
    tcpServer=null;
  }
}

function getServerState() {
  if (tcpServer) {
    return {isConnected: tcpServer.isConnected(),
      addr: tcpServer.addr,
      port: tcpServer.port};
  } else {
    return {isConnected: false};
  }
}
