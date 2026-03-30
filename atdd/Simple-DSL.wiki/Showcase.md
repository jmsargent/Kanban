# 1. Showcase

## Domain overview – financial exchange

Our system is a financial exchange. For those unfamiliar with this domain, an exchange basically contains a number of markets (or financial 'instruments'), each representing something that someone might want to buy or sell. People place orders in a market to buy or sell a quantity of that thing at a given price and we simply match buyers with sellers. There will usually be a gap between the price people are prepared to buy and sell at, which means there will be orders that we can't match together. These orders will sit on the exchange until someone places an order to buy for more or sell for less. The best orders (highest buy price & lowest sell price) sat on the exchange give us the current market price for that market.

## An example test

Let's jump straight into an example. This is a real test from our system. This test could be specified in plain English as follows:

> Given an order that has been placed on the exchange that does not match any other order, when the user cancels the order, then the user should receive an 'order state' event confirming the quantity cancelled.
fgfdgfdgdg
Here's the implementation:

    package com.lmax.exchange.acceptance.test.api;

    import com.lmax.exchange.acceptance.dsl.DslTestCase;
    import org.junit.Before;
    import org.junit.Test;

    public class CancelOrderAcceptanceTest extends DslTestCase
    {
        @Before
        public void beforeEveryTest()
        {
            adminAPI.createInstrument("name: FTSE100");
            registrationAPI.createUser("Bob");

            publicAPI.login("Bob");
        }

        @Test
        public void shouldReceiveCancellationMessageAfterCancellingAnOrder()
        {
            publicAPI.placeOrder("FTSE100", "side: buy", "quantity: 10", "price: 5000", "order: order1");
            publicAPI.cancelOrder("FTSE100", "order1");
            publicAPI.waitForOrderState("order: order1", "cancelledQuantity: 10");
        }

        // other tests...
    }

Hopefully you can see how the implementation maps to its English specification.

We happen to call our setup method 'beforeEveryTest' simply because it reads a little better. In the setup method, we create a financial instrument (a market) called FTSE100, register a new user called Bob, then log Bob into the exchange on our public API.

In the test method, Bob places an order to buy in the FTSE100 market a quantity of 10 units at a price of 5000. We remember his order as 'order1' so we can refer to it later. This is a brand new market and the only person using it is Bob, so his order to buy won't match anyone else's order to sell. His order will remain on the exchange until it matches someone else's order to sell, or Bob cancels it. So next, Bob cancels his order. Finally, we wait for an 'order state' event confirming that a quantity of 10 has been cancelled from the order (i.e. the whole order has been cancelled).


## Test Highlights

Before we get to the mechanics of how the test works, with those wacky String parameters, let's look at the test from a high level.

### It's Java

The test is written in Java, as a normal JUnit test that can be run from an IDE or a build script. There's nothing magical about it. The test assumes that the system under test is already running.


### It's Simple Java

We only use a subset of the Java language syntax. I call it Simple Java. We don't declare fields in the test class, don't use variables or constants, don't use expressions (not even String concatenation), don't use private methods, don't use branching or looping, and so on. We even stick to RuntimeExceptions, just so we don't have to declare checked exceptions on each method. All of our tests look like the one above – more complicated tests just call more methods and use more parameters.

This minimalism is self-imposed and keeps the tests easy on the eye and makes them more accessible to non-technical people. It's also for aesthetics. If you can reduce the amount going on, it makes the tests easier to read, as all you have left is the essence of the test, not a mass of different syntax. You don't have to track which bracket matches up with which, or which context you're in if you're using a fluent-style interface. It's just simple.


### Values are stated explicitly

It can be tempting to use variables, such as extracting the quantity 10 in the above test, but declaring and using the variable would break the visual flow and simplicity of the test. Anyway, it should be obvious that the two 10's are related.

Introducing variables would also tempt you into using expressions, which deviates even further from the idea of keeping things simple. What if you wanted to cancel half of your order? Would you use 'quantity / 2' or keep it simple and just use 5? There's a risk of getting expressions wrong, no matter how simple they are. What if you wanted to verify that the UI renders the value as 5.0 with one decimal place? It's simpler and safer to state values explicitly in tests and use a comment if a calculation needs explanation.


### Domain Specific Language

The test extends DslTestCase, which exposes several fields like publicAPI, adminAPI and tradingUI. These are used to drive different parts of the system under test. Each of these fields exposes various methods like createUser() and placeOrder(), which can take various parameters. The fields, the methods on them and their parameters combine to form a Domain Specific Language, or DSL.

    // adminAPI is a DSL field, createInstrument() is a DSL method and name is a DSL parameter
    adminAPI.createInstrument("name: FTSE100");

The DSL provides an unambiguous, well-defined way of interacting with the system under test. It's a shared vocabulary that can be used to describe those interactions, which can be combined and reused in different ways in different tests.  Pre-defining the fields in DslTestCase reinforces the language because everyone knows what those fields are called. An alternative might be to create the necessary fields in each test, but you'd lose the consistent naming and the shared vocabulary would suffer.

One problem we've encountered is when you want to use more than one instance of a pre-defined field, such as allowing two users to connect on the publicAPI at the same time. Our solution was to keep things a simple as possible and just declare three fields: publicAPI, publicAPI2 and publicAPI3 that should cater for all eventualities. It's not ideal, but it's simple and works. The secondary instances aren't used often, so most of the time you just see publicAPI being used.


### Readability

The primary design goal of SimpleDSL was that tests should be readable by anyone. We should be able to go to a business expert, show them a test, and they should be able to confirm whether the test correctly specifies some desired behaviour. How the DSL is represented (i.e. by String parameters, or a fluent interface, etc) shouldn't get in the way. It shouldn't require lots of explanation, but more than that, it shouldn't break the flow of reading the test. Using Simple Java helps.

We also wanted tests to be machine readable, so you can search and refactor an acceptance test using an IDE. Keeping things simple has had its cost, in that the machine readability of the DSL String parameters has suffered. It is hard to search for usages or rename a DSL String parameter, but at least we can narrow down the search by searching for the DSL method usages. When it came to handling DSL parameters, we decided to sacrifice machine readability for human readability and aesthetics.


### Writeability

Another design goal was that non-developers should be able to write tests. I think it's naive to assume that business analysts and testers aren't capable of writing code, because that's the preserve of 'developers'. In our experience, the use of Simple Java and a well-designed DSL that represents the system as business experts think about it, means practically anyone can write acceptance tests with a little guidance.

At LMAX, our business analysts are happy to write acceptance criteria for new stories as acceptance tests. They mark the tests with an @Ignored annotation until they've been worked on, so they don't break the build. Similarly, testers are happy to flesh out acceptance tests with more edge cases, or document bugs using acceptance tests. If a non-developer needs to invent some DSL, they use their IDE to generate skeleton DSL code that can be implemented by a developer later. When a developer comes to implement the new DSL, they'll usually discuss the intent with the BA / tester and agree any tweaks necessary to make them implementable. The implementation of the DSL should respect the BA / tester's intentions so they can be confident that it accurately reflects the behaviour they desire in the underlying system. Developers frequently work on a story by addressing a set of ignored tests that cover the story's acceptance criteria. Once all the tests pass, the story's done.


### Human / Machine readability balance

There are lots of ways of implementing a DSL. We wanted to strike the right balance between human and machine readability. We want humans to be able to read the tests easily, but when you have thousands of tests, the ability to use tools on them is also essential. Even if you have a handful of tests and your DSL is perfect, you can still bet the business will change tomorrow and you'll have to change your DSL and a whole bunch of existing tests. Without the ability to refactor tests, they'll start to rot, then they'll start to slow you down, then you'll stop using them.

One way of implementing a DSL is by creating a mapping from natural language to Java, as in jBehave. This is obviously very human readable, but not very machine readable. Maybe I'm missing something, but I can't see how jBehave users can search or refactor their tests easily. You don't need to go to the extreme of a natural language representation because BAs and testers are quite capable of writing tests using Simple Java with a good DSL. You might as well capitalise on that and use a DSL representation that's more machine readable.

Another way of implementing a DSL is using a fluent, method chaining style of interface. This is easy to search through and refactor in an IDE, but it isn't as easy to read once you go beyond small tests.

The style we have come up with, with the String parameters, is easy to read, but it's not always easy to refactor and search through. It's easy to search for usages of methods, but parameters can be a pain to refactor. That said, it's a balance that overall, we are prepared to live with. I think a slight compromise on the tooling is worth it for the extra readability, when compared to a fluent interface (the implementation is also a lot simpler than with a fluent interface).


### Test isolation and aliases

Each test sets up its own data (such as the market and the user) to minimise cross-contamination with other running tests and previous test runs. A lot of values in the tests like market names and usernames are actually aliases. When we specify the username 'Bob', the test framework generates a pseudo-unique name like 'Bob-3487394872' and uses that name when interacting with the system under test. In most cases, we don't actually care what the real name is, we just use 'Bob' as an alias (or handle) in the test. Each time we run the test, we will create a new unique Bob. Again, this keeps the tests simple.

We have a convention that if you wrap a value in angle brackets, we will bypass the aliasing mechanism and use the value verbatim. This is useful if we want to test invalid usernames, for example:

    registrationAPI.createUser("name: <username containing spaces>", "expect: invalid username");

### Minimalism

You only need to specify what you care about. Most DSL methods take a variable number of arguments, but a lot of arguments have sensible defaults. For example, when we created the user Bob above, we didn't care what his password or phone number was, so we didn't specify them. And when we logged Bob in, we still didn't care about what his password was, so we didn't specify it and he got logged in with his current password. When we created Bob, we also credited his account with a default amount of money, allowing him to place an order. So there's quite a lot going on in the background when we create a user, but we don't care about that in the test, so we don't distract ourselves with it. Being minimal like this means when things change, we affect less tests.

Also note there is no tear-down method. The test framework keeps track of any resources used in a test and tears them down automatically (such as disconnecting the user from the publicAPI). This keeps the tests simple and avoids problems caused by people forgetting to write teardown methods.


### String parameters

Every method in the DSL boils its parameters down to a series of Strings. Each of these is a name-value pair and sometimes, for certain frequently used parameters, the name is omitted for brevity. A lot of these arguments have sensible defaults, so if you don't pass them in, you assume the defaults. The SimpleDSL library is used to create specifications for these String parameters and to parse them. I'll discuss the SimpleDSL library in the implementation details later.


### Lack of type safety

Some people won't like the lack of type safety in DSL parameters as all values are Strings. In practice, this hasn't been a problem. It just takes a little time to get into the mindset that what you're doing is writing a test using a DSL, not writing a test coding against a Java API. You'll soon find any problems when you run a test after changing it anyway.
