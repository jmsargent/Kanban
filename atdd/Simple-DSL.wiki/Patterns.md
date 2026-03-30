# 5. Patterns

Here's a list of patterns that we've found useful.


## Keywords pattern

Use a keyword instead of an explicit value if you don't care about the value. Since it's a DSL, you can invent keywords to represent certain things. For example, if you want to assert that something is present but don't care what the value is, invent a keyword like PRESENT / ABSENT and use that instead of the actual value.

    publicAPI.waitForOrderState("order: order1", "cancelledQuantity: PRESENT");

## Alias pattern

When you need to use a unique identifier in a test, instead of generating it in the test, pass an alias into the DSL. Generate the unique identifier in the DSL implementation and remember the mapping from the alias to the unique name within the DSL. For example, we use an alias for usernames in tests, but under the hood, we generate a unique username when interacting with the system under test:

    registrationAPI.createUser("Bob"); // Generates a unique name like Bob-82438723 and registers that.
    publicAPI.login("Bob");            // Logs in using Bob-82438723.


## RememberAs pattern

When you need to remember the result of an operation so it can be used later in the test. Instead of assigning the result to a variable, pass in a name and have the DSL operation assign the result to the name passed in. For example, when we place an order, we remember the returned order id using the name "order1" and then cancel that order using that name:

    publicAPI.placeOrder("FTSE100", "side: buy", "price: 5000", "quantity: 10", "order: order1");
    publicAPI.cancelOrder("FTSE100", "order1");

This is a way of implementing assignment and variables in the DSL without having to use the Java syntax for assignment and variables, keeping the DSL simple. There's also the advantage that you could assign to more than one variable at a time, something you can't do in Java.



## Parameter combining

Sometimes, it makes sense to combine several parameters into a single, succinct parameter that follows its own pattern. For example, our system supports a user placing several orders in a single atomic action, known as a mass order:

    fixAPI.placeMassOrder("FTSE100", "bid: 10@38", "bid: 11@39", "ask: 13@44", "ask: 14@45");

Here, the user places an order to buy (bid) a quantity of 10 at a price of 38, another order to buy a quantity of 11 at a price of 39, an order to sell (ask) a quantity of 13 at a price of 44 and another order to sell a quantity of 14 at a price of 45. We can't pass in the prices and quantities as separate parameters because we don't have any way of grouping them together in SimpleDSL. Instead, we group them together using the format 'quantity@price' and pass in values using that syntax. It has the side-benefit of making this type of call very concise.


## Sub-fields for scope

In our DSL, we have a tradingUI DSL field that contains methods for testing the UI in a browser. For example:

    tradingUI.login("Bob");

The UI is quite large, so rather than add all the DSL methods to the tradingUI class itself, we group them into classes representing different parts of the UI. Then we create public fields on tradingUI to provide access to that functionality. For example, we have lots of functionality on the deal ticket (a dialog in our system for placing orders):

    tradingUI.login("Bob");
    tradingUI.showDealTicket("instrument");
    tradingUI.dealTicket.placeOrder("type: limit", "side: buy", "price: 10.0", "quantity: 4.00", "confirmFeedback: false", "timeInForce: GoodTilCancelled");
    tradingUI.dealTicket.checkFeedbackMessage("You have successfully sent a limit order to buy 4.00 contracts at 10.0");

## Negative testing

Since assertions are handled inside the DSL layer, not the test layer, if we expect something to fail, we need to tell the DSL layer to assert that. For example, if we expect creating a user with an invalid username to fail, we need to pass an 'expect failure' flag down to the DSL layer. The DSL method will return normally if the failure happened, or throw a RuntimeException to fail the test if it didn't fail.

    registrationAPI.createUser("<username containing space>", "expect: invalid username");

## Insert alias into message

Sometimes, you want to test a message but the message contains a real value and you only have an alias in the test. For example, you create a user Bob in the test, then want to assert there is an audit log entry in our admin UI containing "Created new user: Bob-12345".

    adminUI.activityLog.waitForActivityEvent("message: Created new user: <user>", "user: Bob");

Here, we use placeholders in the message and pass in other DSL parameters to fill in the placeholders. We look up the real username for the alias 'Bob' and insert that into the expected message, giving us "Created new user: Bob-12345", which we then look for in the admin UI.


## Handling time

We support time in our acceptance tests by creating 'time points', or points in time, then zooming the system forward to a time point. Time points can be derived from the current time (using the keyword `<now>`) or a calculation (startOfTest plus 10 minutes). We can then refer to time points for example, when searching audit logs:

    @Test
    public void viewAuditLogForAccountBetweenKnownTimes()
    {
        dsl.createTimePoint("startOfTest", "<now>");
        dsl.createTimePoint("pointA", "startOfTest plus 10 minutes");
        dsl.createTimePoint("pointB", "pointA plus 10 minutes");
    
        adminAPI.creditAccount("user", "amount: 1.00");
    
        dsl.waitUntil("pointA");  // Our time machine will zoom forward - the test won't just sit here and sleep for 10 minutes!
    
        adminAPI.creditAccount("user", "amount: 2.00");
    
        dsl.waitUntil("pointB");
    
        adminAPI.creditAccount("user", "amount: 3.00");
    
        adminUI.loginAsCustomerServiceAgent();
        adminUI.customerAccounts.viewAuditTrail("user", "startTimePoint: pointA", "endTimePoint: pointB");
        adminUI.customerAccounts.auditEntries.ensureEntryExists("Transfer request [0-9]*, type CREDIT for 2\\.00 in GBP\\. Reason 'Deposit'\\.");
        adminUI.customerAccounts.auditEntries.ensureEntryDoesNotExist("Transfer request [0-9]*, type CREDIT for 3\\.00 in GBP\\. Reason 'Deposit'\\.");
    }
