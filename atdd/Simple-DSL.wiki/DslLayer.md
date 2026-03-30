# 3. DSL Layer

Finally, the bit about those String parameters and the SimpleDSL library I mentioned at the beginning!


## Calling DSL methods

Every DSL method takes its DSL parameters as a variable number of String parameters. Each parameter is a name-value pair separated by "=" or ":", e.g. "name: Bob". If you had "time: 10:45" then the value would be "10:45" as we split on the first separator char.

We distinguish between required and optional DSL parameters. Required parameters must always appear first and in the order specified in the method. As we know the position of each required parameter, the name part of the parameter can be omitted for brevity. Optional parameters can appear in any order after the required parameters and should always contain the name part. For example, the createUser() method takes one required parameter named 'name' and several optional parameters, so these are all valid calls:

```java
// The first parameter is required to be the name parameter:
registrationAPI.createUser("name: Bob");

// As it is required, you can omit the name part:
registrationAPI.createUser("Bob");

// Optional parameters follow the first parameter and can be in any order:
registrationAPI.createUser("name: Bob", "password: P455w0rd", "title: Mr", "dateOfBirth: 1970-01-01", "currency: EUR", "balance: 1000");
registrationAPI.createUser("Bob", "balance: 1000", "currency: USD");
```

Some parameters allow multiple values to be passed in. These can be passed in as a single comma-separated parameter value or multiple parameters with the same name. For example, the activateInstruments() method can take multiple parameters named "instrument":

```java
adminApi.activateInstruments("instrument: FTSE100, DAX");
adminApi.activateInstruments("instrument: FTSE100", "instrument: DAX");
```

## Implementing DSL methods

Here's how createUser() is implemented (this is a very cut-down version of the real code):

```java
package com.lmax.exchange.acceptance.dsl.language;

import java.math.BigDecimal;

import com.lmax.dslparams.DslParams;
import com.lmax.dslparams.OptionalParam;
import com.lmax.dslparams.RequiredParam;
import com.lmax.exchange.acceptance.dsl.support.TestContext;
import com.lmax.exchange.acceptance.framework.driver.SystemDriver;

// DSL class
public class RegistrationAPI
{
    private final SystemDriver systemDriver;
    private final TestContext testContext;

    public RegistrationAPI(final SystemDriver systemDriver, final TestContext testContext)
    {
        this.systemDriver = systemDriver;
        this.testContext = testContext;
    }

    // DSL method
    public void createUser(final String... args)
    {
        final DslParams params = new DslParams(args,
                                               new RequiredParam("name"),
                                               new OptionalParam("password").setDefault("password"),
                                               new OptionalParam("title").setDefault("Mrs"),
                                               new OptionalParam("dateOfBirth").setDefault("1976-06-24"),
                                               new OptionalParam("currency").setDefault("GBP").setAllowedValues("GBP", "EUR", "USD", "AUD"),
                                               // lots more parameters ...
                                               new OptionalParam("balance").setDefault("15000.00"));

        final String alias = params.value("name");
        final String uniqueUsername = testContext.lookupOrCreateUserName(alias);

        final long accountId = registerUser(uniqueUsername, params);

        creditAccount(accountId, params);
    }

    private long registerUser(final String username, final DslParams params)
    {
        final RegistrationUser registrationUser = new RegistrationUser();
        registrationUser.setUsername(username);
        registrationUser.setPassword(params.value("password"));
        registrationUser.setCurrency(params.value("currency"));
        registrationUser.setTitle(params.value("title"));
        registrationUser.setDateOfBirth(params.value("dateOfBirth"));

        final long accountId = systemDriver.getRegistrationDriver().registerUser(registrationUser);

        testContext.storeUserDetails(accountId, registrationUser);
        return accountId;
    }

    private void creditAccount(final long accountId, final DslParams params)
    {
        final BigDecimal amountToCredit = params.valueAsBigDecimal("balance");
        if (amountToCredit != null && amountToCredit.compareTo(BigDecimal.ZERO) != 0)
        {
            systemDriver.getAdminApiDriver().creditAccount(accountId, amountToCredit, "Deposit");
        }
    }
}
```

## DslParams

The first line of a DSL method should always create a DslParams object. This is a parser for the String parameters that uses RequiredParam and OptionalParam objects to specify the valid DSL parameters for this method. It also serves as useful documentation of the DSL parameters.

As you can see, RequiredParams must come before OptionalParams and the order of any RequiredParams specifies the order in which values must be passed in. You can embellish each parameter object to set things like default values or a list of valid values. The embellishments work like the builder pattern – they return the original RequiredParam / OptionalParam object, so they can be chained.

Once you have a DslParams object, you can query it for the values passed in by name, e.g. `params.value("name")` returns "Bob" with leading and trailing spaces trimmed. If there are multiple values passed in for a parameter, you can query the multiple values using `params.values("name")`, which returns a String array. When the DslParams object is constructed, it validates that the args passed into it conform with the specified RequiredParam and OptionalParams. If they don't, it throws an IllegalArgumentException, which will fail the acceptance test.


## RequiredParam and OptionalParam

A RequiredParam specifies that the DSL parameter passed into the DSL method at that position is mandatory. It specifies the name of the parameter, but the name can be omitted for brevity when calling the method as the value can be located by its position. However, if the name is specified, it must match the name in the RequiredParam.

An OptionalParam object specifies an optional parameter that can appear in any position after the required parameters. As it can appear in any position, the name part of the parameter must always be specified.


## Default values

You can specify a default value for an OptionalParam, such as the default password above. This works as you'd expect – if the DSL method is called without a value for that parameter, the default is returned when queried. Defaults cannot be specified on RequiredParam because required parameters must always be specified.


## Enumerating valid values

Both RequiredParam and OptionalParam support enumerating the valid values that can be passed in in that parameter, e.g. the currency parameter above. If the DSL method is called with a parameter value that doesn't match one of the enumerated values, an IllegalArgumentException is thrown.


## Allowing multiple values

Both RequiredParam and OptionalParam also support multiple values being passed in. For example, the activateInstruments() DSL method supports passing multiple instrument names in:

```java
public void activateInstruments(final String... args)
{
    final DslParams dslParams = new DslParams(args, new RequiredParam("instrument").setAllowMultipleValues());
    for (final String instrumentAlias : dslParams.values("instrument"))
    {
        final InstrumentDetails instrumentDetails = testContext.getInstrument(instrumentAlias);
        if (instrumentDetails != null)
        {
            activateInstrument(instrumentDetails);
        }
    }
}
```

This can be called using a comma separated list of values in one parameter or multiple parameters:

```java
adminApi.activateInstruments("FTSE100, DAX");
adminApi.activateInstruments("instrument: FTSE100, DAX");
adminApi.activateInstruments("instrument: FTSE100", "instrument: DAX");
```

Note that you need to be a little careful when calling a method that allows multiple values for a required parameter that is followed by another required parameter. For example, take this specification:

```java
public void dslMethod(final String... args)
{
    final DslParams dslParams = new DslParams(args, 
                                              new RequiredParam("apples").setAllowMultipleValues(),
                                              new RequiredParam("oranges"));
}
```

The following is a valid call. DslParams assigns each argument passed into the method to the 'apples' DSL parameter until it reaches the first parameter that is named and isn't named 'apples', i.e. the first one named 'oranges'. All the rest will be assigned to the oranges DSL parameter as there are no more parameters named anything other than 'oranges'. Basically, required parameters need to appear in order, not necessarily in the exact position. Also, you need to name the first parameter that follows the end of a multiple value sequence in order to break the sequence, even if the next parameter is a required parameter.

```java
dslObject.dslMethod("cox", "apples: braeburn", "apples: bramley, pippin", "dabinett",
                    "oranges: mandarin", "valencia", "oranges: hamlin, gardner", "joppa");
```

## Single RequiredParam shortcut

If you only have one required parameter in a DSL method, there is a shortcut for specifying its name, creating a DslParams object and getting the value out of it:

```java
final String userNameAlias = DslParams.getSingleRequiredParamValue(args, "userName");
```

## SimpleDSL library
    
The open source SimpleDSL library contains DslParams, RequiredParam, OptionalParam and one or two supporting classes. The rest of the code described here is specific to our system and is more of a pattern to copy than something that can be used verbatim. The TestContext contains objects specific to our domain, the SystemDriver contains drivers specific to our system, and the DslTestCase contains DSL fields specific to our system.


## Other details

The chunk of RegistrationAPI code above demonstrates a few other things. The TestContext is used to create a pseudo-unique username from an alias (or look up the username from the alias if it's already been done). The TestContext remembers the mapping so the username can be looked up later. We store other things in the TestContext, like the RegistrationUser object, which we probably use for the password when we log the user in.

RegistrationAPI makes use of more than one type of driver – after creating a user, it can also credit the user with some money using the AdminApiDriver. Strictly speaking, the RegistrationAPI DSL class should only use registration related drivers, but this demonstrates that the DSL object has full access to all the drivers.

We get the RegistrationDriver and AdminApiDriver when we need them, rather than getting them in the RegistrationAPI constructor and storing them in fields. The DslTestCase class will always instantiate the RegistrationAPI class to assign to its registrationAPI DSL field, so the RegistrationAPI DSL class should keep its constructor simple.

We call the first DSL parameter 'name' but use it as an alias. 'name' makes more sense when reading a test; it's an implementation detail that we use the name as an alias rather than literally.