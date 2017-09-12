# go-API-template
An attempt at creating a RESTful API template (built with Go).
- I'm learning Go as a hobby currently (started in January 2017) and absolutely love it.  I'm hoping to start to use Go professionally soon and in order to do so, I want to make sure I'm able to build "industrial strength" REST APIs, in order to mentor my team. The goal of this app is to make an example/template of relational database-backed APIs that have characteristics needed to ensure success in a high volume environment. This is a work in progress - you'll notice most things below are not checked, but I will get there!  Any feedback and/or support are welcome. I'm not an expert and have very thick skin, so please feel free to tell me how bad something is and I'll make it better.

### Critical components of any API (in no particular order)
- [ ] Unit Testing (with reasonably high coverage %)
- [ ] Verbose Code Documentation
- Instrumentation
    - request/response logging (ability to turn on and off logging type based on some type of flag)
        - [ ] logStyle 1: structured (JSON), leveled (debug, error, info, etc.) logging to stdout
        - [ ] logStyle 2: relational database logging (certain data points broken out into standard column datatypes, request/response stored in TEXT/CLOB datatype columns)
    - [ ] API Metrics
    - [ ] Performance Monitoring
    - [ ] Helpful logging
- [ ] Fault tolerant - Proper Error Raising/Handling
- [ ] RESTful service versioning
- [ ] Security/Authentication/Authorization using HTTPS/OAuth2, etc.
- [ ] Properly Vendored dependencies (I have no idea how to do this yet...)
- [ ] Containerized
- [ ] Generated Client examples
- [ ] Extensive API Documentation for Clients of the API (see [twilio](https://www.twilio.com/docs/api/rest), [Uber](https://developer.uber.com/docs/riders/ride-requests/tutorials/api/introduction), [Stripe](https://stripe.com/docs/api/go#intro) and [mailchimp](http://developer.mailchimp.com/documentation/mailchimp/) as good examples)

### Helpful Resources I've used in this app
- JustforFunc
- Go By Example
- 