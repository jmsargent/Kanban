# 4. Miscellaneous notes

## Test organisation

I recommend organising tests around functional areas rather than stories. When a piece of functionality has changed a lot over time, the history of stories that led to the final behaviour becomes irrelevant. What's more important is to have a suite of cohesive tests that cover the current behaviour. Tests for specific features become easier to find and it's easier to check the whether requirements have fallen between the cracks left by different stories. It's enticing to be able to say that each story should be covered by an acceptance test, but that shouldn't be taken literally. Acceptance tests for bugs should also be incorporated into the tests for the functional area, not left in a special bugs package. Grouping tests by area means it's easy to run just the tests for that area if you've just done some work on it.


## Unexpected benefits

We have suffered a lot from intermittent, seemingly random test failures, although we're getting better all the time. Sometimes, they're caused by race conditions in a sloppy test or a poor DSL implementation. (Ours is a highly asynchronous system, so you usually have to wait for some kind of asynchronous acknowledgement of something happening before continuing to the next step in a test). Or they may be caused by an environmental issue, such as one of the test machines going down. But occasionally, we find they're caused by real race conditions or other bugs in our system itself. _These real bugs are ones that we'd never discover if we didn't have so many automated acceptance tests._ They're usually isolated to edge cases that only happen when a certain combination of events happen, so they can take a long time to track down – sometimes months. We discover these bugs as a side-effect of having so many tests.


## Waiting

Our system is highly asynchronous, so in lots of places, we have to wait for something to happen. Sometimes we're explicit about it, such as waiting for a cancelled order state event coming out of the system in the CancelOrderAcceptanceTest. Sometimes it's implicit. For example, when we create a user, we have to wait for the user's details to be propagated to various parts of the exchange before we can make a credit to that user's account.

The important point is that wherever we wait, we hide the waiting in the DSL layer, which keeps the tests simple. Waiting has a timeout, after which the test will fail.


## Levels of abstraction

When inventing new DSL, we try to use the level of abstraction that matches how someone would describe interacting with the system. For example, register a new user, login, place an order, etc:

    registrationAPI.createUser("Bob");
    tradingUI.login("Bob");
    tradingUI.placeOrder("instrument", "side: buy", "quantity: 200", "price: 69");

When writing tests, we use that level of abstraction to set up the test scenario, but then we sometimes want to switch to a lower level of abstraction in order to magnify and test what happens next in more detail. If the differences between the levels of abstraction are small, we might add more parameters to a DSL method to fine-tune how it works in more detail. That's effectively making one method support two levels of abstraction. For example, we have a 'deal ticket' dialog for placing an order. There is a medium level of abstraction DSL method for placing an order on the deal ticket, which usually dismisses the feedback message that pops up after placing the order. But we have added a confirmFeedback parameter to that method to control what it does in more detail, so we can test the feedback message:

    tradingUI.showDealTicket("instrument");
    tradingUI.dealTicket.placeOrder("type: limit", "side: buy", "price: 10.0", "quantity: 4.00", "confirmFeedback: false");
    tradingUI.dealTicket.checkFeedbackMessage("You have successfully sent a limit order to buy 4.00 contracts at 10.0");
    tradingUI.dealTicket.dismissFeedbackMessage();

If the differences between the levels of abstraction are large, we'll usually create one or more new DSL methods. This is most notable in our UI. We'll usually place the more detailed DSL methods on a public sub-field of a DSL field (see the patterns section). For example, we use a low level of abstraction for testing 'stop loss price' feedback:

    tradingUI.dealTicket.enterFieldValues("type: Limit", "price: 60");

    tradingUI.dealTicket.typeIntoField("stopLossOffset: 8");
    tradingUI.dealTicket.checkEstimatedStopLossPrices("68.0 / 52.0");

    tradingUI.dealTicket.typeIntoField("stopLossOffset: BACKSPACE");
    tradingUI.dealTicket.checkEstimatedStopLossPrices("HIDDEN");
