# TODO

## Context
- Update the StartContext methods to use OpenTelemetry (Ref: https://opentelemetry.io/docs/instrumentation/go/getting-started/)
- Provide set up for openTelemety exporting
-- Prometheuse
-- StdOut
-- Elastic Search
-- Azure Monitor? https://pkg.go.dev/github.com/adamko147/opentelemetry-azure-monitor
- Update Context to better support capturing models / attributes
- Update Context to better support capturing events
- Add the User object to the context
- provide some CtxSet and CtxGet options
- Need to have an easy way to setup the exporters / desinations I also like being able to capture a single span to a file / string (like a gilab request process) to let us 
- Talk to mike about our guidelines for creating traces

## Environment / Configuration
See [Configuration](./docs/configuration.md)

## VM API
- Add the create api
- Add the get api
- Add a way to enable auto shutoff behavior

## Auth
- We have a standardized user object
- Create a process that can generate API keys. An API key will act as a user. Not sure if we want to generate a user too? Need to discuss if we take the user to user-token approach or if we want to have a unique API key. 
- Need to discuss 