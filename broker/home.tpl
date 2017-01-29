<html>
    <head>
        <script
                src="https://code.jquery.com/jquery-3.1.1.min.js"
                integrity="sha256-hVVnYaiADRTO2PzUGmuLJr8BLUSjGIZsDYGmIJLv2b8="
                crossorigin="anonymous"></script>

        <script>
            $( document ).ready(function() {
                $("#GetMarketData").on('click', function() {
                    var symbol = $("#symbol").val();

                    $.get("/marketData?symbol=" + symbol, function( data ) {
                        $("#result").text( data );
                    });
                });
            });

        </script>
    </head>
    <body>
        <input id="symbol" type="text" plceholder="Symbol"></input> <input value="Get Market Data" id="GetMarketData" type="button"></input>
        <pre id="result">

        </pre>
    </body>
</html>