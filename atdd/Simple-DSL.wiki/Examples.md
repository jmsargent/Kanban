# 6. Other examples

Here are some more examples, just to give you an idea of how we use our DSL.

```java
public class OrderBookOpenAndCloseNotificationAcceptanceTest extends DslTestCase
{
    @Before
    public void setUp()
    {
        dsl.enableTimeMachine("timeTravelTo: <next weekday>");
        dsl.createTimePoint("name: origin", "value: <now>");
        dsl.createTimePoint("name: open", "value: origin plus 1 weekdayCalendarOpenOffset");
        dsl.createTimePoint("name: close", "value: origin plus 1 weekdayCalendarCloseOffset");
        dsl.createTimePoint("name: nextOpen", "value: open plus 1 weekday");
        dsl.waitUntil("open");

        registrationAPI.createUser("user", "accountType: STANDARD_TRADER");
        registrationAPI.createUser("marketMaker", "accountType: MARKET_MAKER");
        publicAPI.login("user");
        tradingUI.login("user");
        fixAPI.login("marketMaker");

        adminAPI.createInstrument("name: instrumentToClose", "calendar: <WeekdayCalendar>");
        publicAPI.subscribeToOrderBookStatusEvents("instrumentToClose");

        publicAPI.waitForOrderBookStatus("instrumentToClose", "Opened");
    }

    @Test
    public void shouldDisplayDashForBuyAndSellPricesForAnInstrumentThatIsClosedOnInstrumentPanel()
    {
        tradingUI.instrumentSearch.search("instrument: instrumentToClose");

        fixAPI.placeMassOrder("instrumentToClose", "bid: 10@49.0", "ask: 10@51.0");
        tradingUI.allInstruments.list.waitForBestBidAndAskPricesForInstrument("instrument: instrumentToClose", "bid: 49.0", "ask: 51.0");

        dsl.waitUntil("close");
        publicAPI.waitForOrderBookStatus("instrumentToClose", "Closed");

        tradingUI.allInstruments.list.waitForBestBidAndAskPricesForInstrument("instrument: instrumentToClose", "bid: -", "ask: -");

        tradingUI.allInstruments.list.addToWatchlist("instrumentToClose");
        tradingUI.watchlist.list.checkInstrumentList("instrument: instrumentToClose");
        tradingUI.watchlist.list.waitForBestBidAndAskPricesForInstrument("instrument: instrumentToClose", "bid: -", "ask: -");
    }

    // ...
}
```

In this test, we set up a market that has orders in it and view the current best buy and sell prices in the UI. When the market closes at the end of the day, we check that the best buy and sell prices are removed from the UI.

The above test shows off the time machine. Our system abstracts over getting the current time, and the time machine installs a dummy clock that can be controlled from the tests.

The above test also shows that we sometimes group related DSL calls together, such as in tradingUI.allInstruments.list (our UI has a panel called 'all instruments' and that can be viewed as a list or series of widgets). allInstruments and list are public fields. The list object is always present, but each of the methods on it (like addToWatchlist) lazily create the underlying driver, so we don't initialise the driver (and Selenium) unless we actually use it.
