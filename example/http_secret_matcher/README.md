Example application for instafasthttp
=======================================

This example demonstrates how to configure secret matchers using the Instana Go tracer. It showcases the use of two types of matchers combined into a single MultiMatcher:
- `EqualsIgnoreCaseMatcher` for the string `"key"`
- `ContainsIgnoreCaseMatcher` for the string `"pass"`


Running the application
------------------------
- Ensure an Instana host agent is running to collect traces (or use agentless tracing if preferred).
- Build the application using:
```bash
go build -o server .
```
- Run the executable:
```bash
./server
```

## You can test the server by visiting:
- [localhost:7070/endpoint?keys=123&password=456](http://localhost:7070/endpoint?keys=123&password=456)

## Expected Result
The request includes two query parameters: `keys` and `password`.
    - The keys parameter will not be masked because the matcher is looking for the exact string "key".
    - The password parameter will be masked because it contains the substring "pass", which matches the configured matcher.
You can verify this behaviour on the Instana dashboard.
