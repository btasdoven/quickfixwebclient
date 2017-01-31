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
                    var side = $("#side :selected").val();

                    var uri = "/orderSingle" +
                        "?symbol=" + symbol +
                        "&quantity=" + quantity +
                        "&limit=" + limit +
                        "&side=" + side;

                    $.ajax(uri, {
                        method: "GET",
                        xhrFields: { withCredentials: true },
                        crossDomain: true,
                        success: function( data ) {
                            $("#result2").text(data);

                            var request_time = new Date().getTime() - start_time;
                            $("#timerbox").text("Request took " + request_time + "ms to complete.")
                        }
                    });
                });

                $("#GetOrders").on('click', function() {
                    var start_time = new Date().getTime();

                    $.ajax("/orders", {
                        method: "GET",
                        xhrFields: { withCredentials: true },
                        crossDomain: true,
                        success: function( data ) {
                            $("#result3").text(data);

                            var request_time = new Date().getTime() - start_time;
                            $("#timerbox").text("Request took " + request_time + "ms to complete.")
                        }
                    });
                });

            });

        </script>
    </head>
    <body>
        <pre id="timerbox">
        </pre>

        <hr/>

        <input id="symbol" type="text" placeholder="Symbol" value="MSFT"></input>
        <input value="Get Market Data" id="GetMarketData" type="button"></input>
        <pre id="result">



        </pre>

        <hr/>

        <input id="symbol2" type="text" placeholder="Symbol" value="MSFT"></input>
        <input id="order_quantity" type="number" placeholder="Order Quantity"></input>
        <input id="order_limit" type="number" placeholder="Order Limit"></input>
        <select id="side">
            <option value="1">BUY</option>
            <option value="2">SELL</option>
        </select>
        <input value="Order Single" id="OrderSingle" type="button"></input>
        <pre id="result2">
        </pre>

        <hr/>

        <input value="Get Orders" id="GetOrders" type="button"></input>
        <pre id="result3">

        </pre>
    </body>
</html>