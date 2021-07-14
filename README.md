# Caddy Conditional Logging

> Hey Caddy, log only if ...

This plugin implements a logging encoder that let's you log depending on conditions.

Conditions can be express through a simple expression language.

## Module

The module name is `if`.

Its syntax is:

```caddyfile
if {
    <field> <operator=eq|ne|sw> <value>
    ...
} [<other encoder>]
```

This Caddy module will wrap the `<other encoder>` (eg. `json`, or `console`) and log if at least one of the conditions is met.

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

Log depending on the enviroments (console?) to stdout if the request's method is "GET".

```caddyfile
log {
  output stdout
  format if {
      request>method eq GET
  }
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

## Try it out

From the root directoy of this project, run:

```console
xcaddy run
```

Then open <https://localhost:2015>, go on existing and non-existing pages, and observe the access logs.

---

[![Analytics](https://ga-beacon.appspot.com/UA-49657176-1/caddy-conditional-logging?flat)](https://github.com/igrigorik/ga-beacon)
