(function (window) {
    // This will ignore if this script is inserted again
    if (window._httpserve_listener) { return }
    window._httpserve_listener = 1

    // console overrides
    const consoleOriginal = {}
    const consoleOverrides = {}


    // HIJACK console.log and send it to server
    // Send log to server for development
    for (let c in console) {
        if (typeof console[c] !== 'function') { continue }
        consoleOriginal[c] = console[c]
        console[c] = function () {
            args = [...arguments]
            consoleOverrides[c] &&  consoleOverrides[c].apply(this,args)
            consoleOriginal[c].apply(this, args)
        }
    }

    var ws = null
    var loc = 'ws://' + window.location.host + '/.httpServe/_reload'
    function connect(loc) {
        ws = new window.WebSocket(loc)
        ws.onopen = function() {
            // Grab files to send to watcher
            ws.send(JSON.stringify({
                "op": "watch", 
                "value": listResources(),
            }))

            // Error to server
            window.onerror = function() {
                ws.send(JSON.stringify({
                    "op": "error", 
                    "value": [...Array.from(arguments)]
                }))
                return false
            }
            for (let c in consoleOriginal){
                consoleOverrides[c] = function(){
                    ws.send(JSON.stringify({
                        "op": "log", 
                        "type": c, 
                        "value": [...Array.from(arguments)],
                    }))
                }
            }
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

function listResources(){
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
    return fileList
}


