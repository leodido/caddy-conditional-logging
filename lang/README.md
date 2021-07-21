# Simple bool expressions

A simple yet awesome bool expression language.

## Expression

An expression is - usually - in the form of `<lhs> <operator> <rhs>`.

## Logical operators

- `&&`: AND
- `||`: OR

## Comparison operators

- `==`: equals to
- `!=`: not equals to

## Fields

The `<lhs>` and `<rhs>` can be **fields**, even both at the same time, or **other expressions**

For example `status == status` is a valid expression.

The **field syntax** is as per [buger/jsonparser](https://github.com/buger/jsonparser).

So, you can assert conditions also on nested fields!

Let's say you want to log if the user agent starts starts with some value...

You'd need to traverse the request object, its headers child (another object), and its "User-Agent" child (array).

Sounds difficult. But, it isn't! Express this field as: `request>headers>User-Agent>[0]`.

## Values

The `<lhs>` and `<rhs>` can be **values**, of course.

The expressions support the following types for the values: boolean, (raw) string, and numbers (int64, and float64).

The language provides two boolean constants (case-insensitive): `true` and `false`.

### Constants

This means that `<lhs>` and `<rhs>` (whether are they fields or constants) need to be of one of those types.

## Operator precedence

In order of evaluation:

1. `()`
2. `||`
3. `&&`
4. `==`, `!=`
