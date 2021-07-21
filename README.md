# Caddy Conditional Logging

> Hey Caddy, please log only if ...

This plugin implements a logging **encoder** that let's you **log depending on conditions**.

Conditions can be express through a simple expression language.

## Module

The **module name** is `if`.

Its syntax is:

```caddyfile
if "<expression>" [<encoder>]
```

This Caddy module logs as the `<encoder>` demands if at least one of the expressions is met.

The `<expression>` must be enclosed in double quotes.

The supported encoders are:

- [`json`](https://caddyserver.com/docs/caddyfile/directives/log#json)
- [`console`](https://caddyserver.com/docs/caddyfile/directives/log#console)
- [`jsonselect`](https://github.com/leodido/caddy-jsonselect-encoder)

When no `<encoder>` is specified, a default encoder (`console` or `json`) is automatically set up depending on the environments.

### Expressions

The [language](./lang) supports simple boolean expressions.

An expression is - usually - in the form of `<lhs> <operator> <rhs>`. But you can compose and nest them!

Take a look at the [language documentation](./lang/README.md) for more information.

## Caddyfile

Log JSON to stdout if the status starts with a 4 (eg., 404).

```caddyfile
log {
  output stdout
  format if "status ~~ `^4`" json
}
```

Log to stdout in console format if the request's method is "GET".

```caddyfile
log {
  output stdout
  format if "request>method == `GET`" console
}
```

Log JSON to stdout if at least one of the conditions match.

```caddyfile
log {
  output stdout
  format if "status ~~ `^4` || status ~~ `^5` || request>uri == `/`" json
}
```

Log JSON to stdout if the visit is from a Mozilla browser.

```caddyfile
log {
  output stdout
  format if "request>headers>User-Agent>[0] ~~ `Gecko`" json
}
```

Log a JSON containing only the timestamp, the logger name, and the duration
for responses with HTTP status equal to 200.

```caddyfile
log {
  format if "status == 200" jsonselect "{ts} {logger} {duration}"
}
```

This outputs a nice JSON like the following one:

```json
{"ts":1626440165.351731,"logger":"http.log.access.log0","duration":0.000198292}
```

Do you wanna log [Stackdriver](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry) entries only for 4** response status codes?

Let's do it!

Change the level and time format, and also change the key names for the resulting JSON.

```caddyfile
log {
  format if "status ~~ `^4`" jsonselect "{severity:level} {timestamp:ts} {logName:logger} {httpRequest>requestMethod:request>method} {httpRequest>protocol:request>proto} {httpRequest>status:status} {httpRequest>responseSize:size} {httpRequest>userAgent:request>headers>User-Agent>[0]} {httpRequest>requestUrl:request>uri}" {
    level_format "upper"
    time_format "rfc3339_nano"
  }
}
```

This outputs:

```json
{"severity":"INFO","timestamp":"2021-07-19T15:44:44.077586Z","logName":"http.log.access.log0","httpRequest":{"requestMethod":"GET","protocol":"HTTP/2.0","status":200,"responseSize":11348,"userAgent":"Mozilla/5.0 ...","requestUrl":"/leo"}}
```

## Try it out

From the root directoy of this project, run:

```console
xcaddy run
```

Then open <https://localhost:2015>, go on existing and non-existing pages, and observe the access logs.

To install xcaddy in case you need to, run:

```console
go get -u github.com/caddyserver/xcaddy/cmd/xcaddy
```

## Build

To build [Caddy](https://github.com/caddyserver/caddy) with this module in, execute:

```console
xcaddy build --with github.com/leodido/caddy-conditional-logging
```

---

[![Analytics](https://ga-beacon.appspot.com/UA-49657176-1/caddy-conditional-logging?flat)](https://github.com/igrigorik/ga-beacon)
