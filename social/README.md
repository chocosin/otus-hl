# Social network prototype

## Interesting paths:

`/` - index, redirects to user page if logged in

`/last` - last registered users

`/user/qqq` - page of user with username qqq

Only available for non-registered users:

`/signup`

`/login`

## Authentication
Authentication is implemented by storing uuid token in cookies.

Token is generated at login time.
