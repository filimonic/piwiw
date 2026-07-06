# PIWIW

A tiny dumb microservice to proxy Ollama Chat API to OpenAI chat completion API.

## Configuration:

`piwiw` is configured using environment variables: 

| Var name                      | Type    | Default    | Description
| ----------------------------- | ------- | ---------- | -----------
| SERVER_PORT                   | integer | 11434      | HTTP server port piwiw listens on
| _PROXY_PORT_                  | integer | 11434      | Alias to `SERVER_PORT`. Ignored if `SERVER_PORT` is set. _For [eyalrot/ollama_openai](https://github.com/eyalrot/ollama_openai) compatibility_
| OPENAI_API_BASE_URL           | string  | _required_ | HTTP(S) base URL of OpenAI API endpoint, ex `https://api.example.com/v1`
| SKIP_TLS_VERIFY               | bool    | `false`    | Skip OpenAI API certificate verification
| OPENAI_API_KEY                | string  | _required_ | OpenAI API endpoint bearer token, ex `sk-12345`
| OPENAI_API_CHAT_FORCED_PARAMS | json    | `{}`       | Params that will be added (replaced) to OpenAI Chat Completion API request, ex `{"model":"my-model"}` will force model to be used.
| OPENAI_API_CHAT_FORCED_PARAMS_B64 | base64 | _(none)_ | Same as `OPENAI_API_CHAT_FORCED_PARAMS`, but base64-encoded. Useful when the plain JSON value gets mangled by intermediate YAML/templating tools. Mutually exclusive with `OPENAI_API_CHAT_FORCED_PARAMS`.
| REQUEST_TIMEOUT               | seconds | 180        | Number of seconds `piwiw` will wait for response from OpenAI API for single request
| MAX_RETRIES                   | integer | 3          | For failed requests by `REQUEST_TIMEOUT` or with `5xx` response code, `piwiw` will retry request `MAX_RETRIES` times. 
| RETRY_DELAY                   | seconds | 300        | Request retry will be delayed for `RETRY_DELAY`.
| EMPTY_CONTENT_TEXT            | text    | `""`       | If OpenAI respone contains message with null or empty content, this text will be passed back to ollama client.
| TRACE_FOLDER_PATH             | string  | _(none)_   | If set, `piwiw` traces every request/response pair as JSON files into a per-request folder under this path. Tracing is disabled when unset.
| TRACE_KEEP_HOURS              | integer | 2160       | Trace folders older than this many hours (based on the folder name, checked hourly) are deleted. Only relevant when `TRACE_FOLDER_PATH` is set.

## Logging

Every request gets a short random request ID (Crockford Base32). Log lines are prefixed with `[REQUEST_ID]` followed by a one-letter level (`I` info, `W` warning, `E` error), e.g.:

```
[FNJB2S57ZQ] I New incoming request from [::1]:54040
[FNJB2S57ZQ] I OpenAI request started (attempt 1/4)
[FNJB2S57ZQ] I OpenAI request completed (status 200)
[FNJB2S57ZQ] I Incoming request responded and completed (status 200)
```

## Tracing

When `TRACE_FOLDER_PATH` is set, each incoming request creates a folder named `<UTC timestamp>.<request_id>` (e.g. `20260704T153000Z.FNJB2S57ZQ`) containing:

- `1_ollama_request.incoming.json` — the incoming Ollama request
- `2_openai_request.outgoing.json` — the outgoing OpenAI request
- `3_openai_response.incoming.json` — the OpenAI response
- `4_ollama_response.outgoing.json` — the Ollama response sent back to the client

Trace folders older than `TRACE_KEEP_HOURS` are deleted once an hour.
