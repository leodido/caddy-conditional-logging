# Caddy Conditional Logging

> Hey Caddy, please log only if ...

This plugin implements a logging **encoder** that let's you **log depending on conditions**.

Conditions can be express through a simple expression language.

## Module

The **module name** is `if`.

Its syntax is:

```caddyfile
if {
    <expression>
    ...
} [<encoder>]
```

This Caddy module logs as the `<encoder>` demands if at least one of the expressions is met.

The supported encoders are:

- [`json`](https://caddyserver.com/docs/caddyfile/directives/log#json)
- [`console`](https://caddyserver.com/docs/caddyfile/directives/log#console)
- [`jsonselect`](https://github.com/leodido/caddy-jsonselect-encoder)

When no `<encoder>` is specified, a default encoder (`console` or `json`) is automatically set up depending on the environments.

### Expressions

Expressions have the following syntax: `<field> <operator> <value>`.

The supported operators are:

- `eq`: equals to
- `ne`: not equals to
- `sw`: starts with

The **field syntax** is as per [buger/jsonparser](https://github.com/buger/jsonparser).

So, you can assert conditions also on nested fields!

Let's say you want to log if the user agent starts starts with some value...

You'd need to traverse the request object, its headers child (another object), and its "User-Agent" child (array).

Sounds difficult. But, it isn't! Express this field as: `request>headers>User-Agent>[0]`.

## Caddyfile

Log JSON to stdout if the status starts with a 4 (eg., 404).

```caddyfile
log {
  output stdout
  format if {
      status sw 4
  } json
}
```

Log to stdout in console format if the request's method is "GET".

```caddyfile
log {
  output stdout
  format if {
      request>method eq GET
  } console
}
```

Log JSON to stdout if at least one of the conditions match.

Notice this means that condistions are in OR.

```caddyfile
log {
  output stdout
  format if {
      status sw 4
      status sw 5
      request>uri eq "/"
  } json
}
```

Log JSON to stdout if the visit is from a Mozilla browser.

```caddyfile
log {
  output stdout
  format if {
      request>headers>User-Agent>[0] sw Mozilla
  } json
}
```

Log a JSON containing only the timestamp, the logger name, and the duration
for responses with HTTP status equal to 200.

```caddyfile
log {
  format if {
      status eq 200
  } jsonselect "{ts} {logger} {duration}"
}
```

This outputs a nice JSON like the following one:

```json
{"ts":1626440165.351731,"logger":"http.log.access.log0","duration":0.000198292}
```

Do you wanna log Stackdriver entries only for 4** response status codes?

Let's do it!

Change the level and time format, and also change the key names for the resulting JSON.

```caddyfile
log {
  format if {
      status sw 400
  } jsonselect "{severity} {timestamp} {logName}" {
    level_key "severity"
    level_format "upper"
    time_key "timestamp"
    time_format "rfc3339"
    name_key "logName"
  }
}
```

This outputs:

```json
{"severity":"ERROR","timestamp":"2021-07-16T12:55:10Z","logName":"http.log.access.log0"}
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
