<html>
    <head>
        <script
                src="https://code.jquery.com/jquery-3.1.1.min.js"
                integrity="sha256-hVVnYaiADRTO2PzUGmuLJr8BLUSjGIZsDYGmIJLv2b8="
                crossorigin="anonymous"></script>

        <script>
            $( document ).ready(function() {
                $("#GetMarketData").on('click', function() {
                    var start_time = new Date().getTime();

                    var symbol = $("#symbol").val();
                    $.get("/marketData?symbol=" + symbol, function( data ) {
                        $("#result").text(data);

                        var request_time = new Date().getTime() - start_time;
                        $("#timerbox").text("Request took " + request_time + "ms to complete.")
                    });
                });

                $("#OrderSingle").on('click', function() {
                    var start_time = new Date().getTime();

                    var symbol = $("#symbol2").val();
                    var quantity = $("#order_quantity").val();
                    var limit = $("#order_limit").val();

                    var uri = "/orderSingle" +
                        "?symbol=" + symbol +
                        "&quantity=" + quantity +
                        "&limit=" + limit;

                    $.get(uri, function( data ) {
                        $("#result2").text(data);

                        var request_time = new Date().getTime() - start_time;
                        $("#timerbox").text("Request took " + request_time + "ms to complete.")
                    });
                });
            });

        </script>
    </head>
    <body>
        <pre id="timerbox"></pre>
        <input id="symbol" type="text" placeholder="Symbol" value="MSFT"></input>
        <input value="Get Market Data" id="GetMarketData" type="button"></input>
        <pre id="result">

        </pre>

        <hr/>

        <input id="symbol2" type="text" placeholder="Symbol" value="MSFT"></input>
        <input id="order_quantity" type="number" placeholder="Order Quantity"></input>
        <input id="order_limit" type="number" placeholder="Order Limit"></input>
        <input value="Order Single" id="OrderSingle" type="button"></input>
        <pre id="result2">

        </pre>
    </body>
</html>