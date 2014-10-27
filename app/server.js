// Functionality dealing with server-level things

var maxScrollback = 500;

function Server(network, baseURL, authtok, messageHandler) {
  function handleEventSourceLine(line) {
    var lhs = line.split(" :", 1)[0]
  	var parts = lhs.split(' ')
  	var timestamp = new Date(parts[0] * 1000);
    var fullSender = parts[1];
  	var command = parts[2];
  	var sender = parts[3];
  	var forum = parts[4];
  	var args = parts.slice(5);
  	var txt = line.substr(lhs.length + 2);

  	messageHandler(timestamp, fullSender, command, sender, forum, args, txt);
  }

  function handleEventSourceMessage(oEvent) {
    msgs = oEvent.data.split("\n");

    var first = Math.max(0, msgs.length - maxScrollback);
    for (var i = first; i < msgs.length; i += 1) {
      handleEventSourceLine(msgs[i]);
    }
  }

  function handleEventSourceError(oEvent) {
    timestamp = new Date();
    messageHandler(timestamp, null, "ERROR", null, null, [], null);
  }

  function send(target, text) {
    function handleError(oEvent) {
      console.log("XXX: That didn't work out.", target, text)
    }

    var form = new FormData();
    form.append("type", "command");
    form.append("auth", authtok);
    form.append("network", network);
    form.append("target", target);
    form.append("text", text);
    console.log(form);

    var oReq = new XMLHttpRequest();
    oReq.addEventListener("error", handleError);
		oReq.open("POST", baseURL, true);
		oReq.send(form);
  }

  this.send = send;

  var pullURL = baseURL + "?network=" + encodeURIComponent(network) + "&auth=" + encodeURIComponent(authtok);
  var eventSource = new EventSource(pullURL);
  eventSource.addEventListener("message", handleEventSourceMessage);
  eventSource.addEventListener("error", handleEventSourceError);
}