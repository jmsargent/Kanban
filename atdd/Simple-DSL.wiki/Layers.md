# 2. Layers

Our test code is organised into three layers: acceptance tests, DSL and drivers.

    Acceptance Tests
    ------------------------
    Domain Specific Language
    ------------------------
    Drivers

The acceptance test layer obviously contains the tests and only uses the DSL layer. This is the layer that non-developers are mainly interested in.

The DSL layer provides the language for the tests. It is responsible for parsing the DSL String parameters and interacting with the driver layer. Some DSL methods are quite simple and just call a driver method. Others are a lot more complicated and may make multiple calls on multiple drivers.

The driver layer contains all the code for driving the system under test. For example, a UI driver may use Selenium to drive our UI in a browser, or an HTTP driver may be used to drive our HTTP API.  The DSL layer isn't the only user of the driver layer – the driver layer is also used by our performance tests and other tools.


## Test layer

There isn't much to say about this layer in terms of its implementation. We have a package for this layer named "test" and all our acceptance tests are organised into sub-packages below it. Every class in these packages is a test and looks similar to CancelOrderAcceptanceTest above.



## DSL layer

We have a package named "dsl" for this layer, below which we have sub-packages for the different parts of the DSL, such as UI, API, etc. The root class in this layer is DslTestCase, which all acceptance test classes subclass. This class exposes the fields that make up the DSL, such as publicAPI, tradingUI, etc. Here's a cut-down version of DslTestCase:

    package com.lmax.exchange.acceptance.dsl;
    
    import java.util.Collection;
    import com.lmax.exchange.acceptance.dsl.language.AdminAPI;
    import com.lmax.exchange.acceptance.dsl.language.RegistrationAPI;
    import com.lmax.exchange.acceptance.dsl.language.ui.TradingUI;
    import com.lmax.exchange.acceptance.driver.SystemDriver;
    
    import org.junit.After;
    import org.junit.AfterClass;
    import org.junit.BeforeClass;
    
    public class DslTestCase
    {
        private final SystemDriver systemDriver = new SystemDriver();
        private final TestContext testContext = new TestContext();
    
        // DSL fields for use in acceptance tests
        protected final AdminAPI adminAPI = new AdminAPI(systemDriver, testContext);
        protected final RegistrationAPI registrationAPI = new RegistrationAPI(systemDriver, testContext);
        protected final TradingUi tradingUI = new TradingUI(systemDriver, testContext);
    
        @BeforeClass
        public static void startUp() throws Exception
        {
            SystemDriver.staticStartUp();
        }
    
        @AfterClass
        public static void shutDown() throws Exception
        {
            SystemDriver.staticShutDown();
        }
    
        @After
        public final void tearDown() throws Exception
        {
            systemDriver.tearDown();
        }
    }

Note the SystemDriver and TestContext classes. The SystemDriver is from the driver layer and provides access to all the drivers. When the registrationAPI DSL field is used in an acceptance test, the registrationAPI object gets a driver for the registration API from the SystemDriver. Note that it does this lazily – DslTestCase always instantiates the registrationAPI object, but if a test doesn't use registrationAPI, it shouldn't initialise a registration API driver. This is particularly important for UI drivers as we don't want the overhead of initialising Selenium and opening a browser unless we're really testing the UI.

The TestContext is like a shared whiteboard that various parts of the DSL can use to share information. For example, when we create a user in the registrationAPI, we generate a real name for the alias used in the test and store the real name and alias in the TestContext. When we login to the PublicAPI, the publicAPI object takes the alias passed into the login method and looks up the real name in the TestContext. Each time you run a test, JUnit creates a new instance of your test class, so you get a new instance of TestContext in DslTestCase, so test data is isolated between tests. TestContext starts out empty and builds up as the test progresses.

The DSL layer is where we use the SimpleDSL library to parse the String parameters. We'll explore DSL methods and String parameter handling later.



## Driver layer

We have a package named "driver" for this layer, below which are sub-packages for the different drivers. The root class in this package is SystemDriver. Each time you run a test, you get a new SystemDriver in DslTestCase. SystemDriver knows how to create each of the drivers for driving the different parts of the system under test. SystemDriver keeps track of every driver it creates. At the end of a test, DslTestCase.tearDown() calls systemDriver.tearDown(), which tears down any drivers that it's created during a test. The DSL layer creates drivers on demand, so we only incur the cost of initialising the drivers actually used.
