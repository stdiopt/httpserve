(function (window) {
  // This will ignore if this script is inserted again
  if (window._httpserve_listener){
    return
  }
  window._httpserve_listener = 1

  var ws = null
  var loc = 'ws://' + window.location.host + '/.httpServe/_reload'

  function connect(loc) {
    ws = new window.WebSocket(loc)

    ws.onopen = function() {
      // Grab files to send to watcher
      var fileList = []
      fileList.push(window.location.pathname)
      // Load assets too
      var elList = document.querySelectorAll('link[href]')
      for (var i =0; i< elList.length; i++ ) {
        var src = elList[i].getAttribute('href')
        if (src.startsWith('/.httpServe')) {
          continue
        }
        if (src.startsWith("data:")){
          continue
        }
        let toWatch = window.location.pathname
        toWatch = toWatch.substring(0, toWatch.lastIndexOf('/'))
        toWatch += '/' + src
        fileList.push(toWatch)
      }
      // Find all src and request a watch too
      var elList = document.querySelectorAll('img[src]')
      for (var i = 0; i < elList.length; i++) {
        var src = elList[i].getAttribute('src')
        if (src.startsWith("/.httpServe")) {
          continue
        }
        if (src.startsWith("data:")){
          continue
        }
        let toWatch = window.location.pathname
        toWatch = toWatch.substring(0, toWatch.lastIndexOf('/'))
        toWatch += '/' + src
        fileList.push(toWatch)
      }
      ws.send(JSON.stringify(fileList))
    }

    ws.onmessage = function(ev) {
      if (JSON.parse(ev.data) === "reload") {
        window.location.reload()
      }
    }
    // Reconnect either on error or close
    ws.onclose  = function(e)  {
      setTimeout(() => connect(loc),3000)
    }
  }
  connect(loc)
})(window)
