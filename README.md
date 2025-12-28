# Coderun
Is a platform, that allows you to run small codes in an isolated environment!

Its back includes two main services: single sign out and executor, as well as additional ones in the form of redis, mongo and clickhouse.

Authorization on the platform is implemented using access and refresh tokens using rsa encryption.

## Single Sign Out
A service that manages access and refresh tokens, and is also responsible for creating and getting new users. It contains the following methods:
- `Register` - register new users, save them to the database, and generate a pair of tokens
- `Login` - register new users, save them to the database, and generate a pair of tokens
- `Refresh` - updating the access token and generating a new refresh (with the deletion of the old one)
- `UpdateUser` - updating the username of user
- `Get...` - get user by id or jwt-token

## Xcutr
A service that runs code in an isolated environment and streams the result. It contains the following method:
- `Execute` - code execution and log translation

