{{define "main"}}
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>WS Client</title>
    <style>
        * {margin: 0; padding: 10px; font-size: 18px;}
        .wait {color: #ccc;}
        .ok {color: #00cc00;}
        .err {color: #cc0000;}
    </style>
</head>
<body>
    <div>Status: <span id="status" class="wait">Connecting...</span></div>
    <div>
        <button id="start">Start</button>
        <button id="stop">Stop</button> 
    </div>
    <div>Number: <span id="number"></span></div>
    <script>
        window.onload = function() {
            var status = document.getElementById("status");
            var start = document.getElementById("start");
            var stop = document.getElementById("stop");
            var number = document.getElementById("number");
            var ws = new WebSocket("ws://127.0.0.1:8080/ws");
            ws.onopen = function(e) {
                status.className = "ok";
                status.innerHTML = "Connected";
            }
            ws.onmessage = function(e) {
                number.innerHTML = JSON.parse(e.data).number;
            };
            ws.onclose = function(e) {
                status.className = "err";
                status.innerHTML = "Disconnected";
            };
            start.onclick = function(e) {
                e.PreventDefault;
                ws.send('{"op": "start"}');
            };
            stop.onclick = function(e) {
                e.PreventDefault;
                ws.send('{"op": "stop"}');
            };
        };
</script>
</body>
</html>
{{end}}