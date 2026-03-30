# Simple DSL for Java acceptance tests

SimpleDSL is a style for writing acceptance tests that aims to balance human and machine readability.  By that, we mean people (developers and non-developers alike) can easily read and understand an acceptance test, and developer IDEs can understand enough of an acceptance test to support searching, refactoring, name completion, etc.  Although we provide a small library of code, SimpleDSL is more of a pattern to follow as most of the code will be specific to your system – the system under test.

SimpleDSL is something I developed and introduced to my company, [LMAX Exchange](http://www.lmax.com/), about 4 years ago.  We take automated acceptance testing seriously at LMAX.  We have thousands of them. Literally!  The last time I checked, we had nearly 6,000 acceptance tests, covering a wide range of areas across our system.  The single build machine that used to run all the tests has grown into a farm of around 30-40 machines, and a run through the entire suite currently takes around 35 minutes.  I'm pleased to see that SimpleDSL has stood the test of time.  The tests we write today follow exactly the same pattern as the first tests we wrote 4 years ago, which gives an indication of how successful this approach has been.

I'm not claiming SimpleDSL is perfect or better than other DSL implementations, but it has worked very well for us and continues to give us a huge amount of value, so for that reason, I think it's worth sharing.  Don't let the simplistic appearance of the tests fool you – I think that's exactly why they're successful.  This simplicity hasn't restricted what we've been able to test.  Our system is a highly complex, highly asynchronous financial exchange and we have tests covering all aspects of business functionality.  We test interactions with (stubbed) third-party systems.  We test disaster recovery, where we kill primary services and test failing over to secondaries.  We even have tests that use a 'time machine' to whisk our system forward in time to test things like end-of-day procedures and promotions timing out.

Writing automated acceptance tests is an integral part of our process, not a sideline that some people do some of the time.  Whenever we add new functionality or fix a bug, we'll write unit tests for the detailed code changes and acceptance tests for the externally observable change in behaviour.  Making the acceptance tests go green for a new story or a bug is a strong indicator that the work is complete.  We'll still do manual testing of course and any new issues found will typically be addressed by writing even more acceptance tests.  We can do this because our acceptance tests have found a sweet-spot where they're light-weight enough that we've been able to add and maintain them as an ongoing part of our process, and powerful enough that they add a huge amount of value.

## What next?

1. [Showcase](wiki/Showcase) - example with high level overview
1. [Layers](wiki/Layers) - brief description of design layers in an acceptance test
1. [DSL Layer](wiki/DslLayer) - detailed description of DSL layer
1. [Miscellaneous notes](wiki/MiscNotes) - some notes about things we've learnt
1. [Patterns](wiki/Patterns) - descriptions of a few patterns we've used
1. [Examples](wiki/Examples) - further examples