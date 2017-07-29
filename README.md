# go-API-template
An attempt at creating a RESTful API template (built with Go)

### Critical components of any API (in no particular order)
- [ ] Extensive API Documentation for Clients of the API (see [twilio](https://www.twilio.com/docs/api/rest), [Uber](https://developer.uber.com/docs/riders/ride-requests/tutorials/api/introduction) and [mailchimp](http://developer.mailchimp.com/documentation/mailchimp/) as good examples)
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