# Caddy Conditional Logging

> Hey Caddy, log only if ...

This plugin implements a logging encoder that let's you **log depending on conditions**.

Conditions can be express through a simple expression language.

## Module

The **module name** is `if`.

Its **syntax** is:

```caddyfile
if {
    <field> <operator=eq|ne|sw> <value>
    ...
} [<other encoder>]
```

This Caddy module will wrap the `<other encoder>` (eg. `json`, or `console`) and log if at least one of the conditions is met.

Notice the module accepts also nested fields.
The **field syntax** is as per [buger/jsonparser](https://github.com/buger/jsonparser).

Let's say you want to log if the user agent starts starts with some value...

You'd need to traverse the request object, its headers child (another object), and its "User-Agent" child (array).

Sounds difficult. It isn't! Express this field as: `request>headers>User-Agent>[0]`.

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
