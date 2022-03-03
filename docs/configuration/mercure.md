## Mercure

For information on what Mercure is, please see the [introduction to Mercure](https://mercure.rocks/docs/mercure) in Mercure documentation.

Mercure can be enabled on `magento2`, `laravel` and on `symfony` env types by changing the following to the project's `.env` file (or exporting them to environment variables prior to starting the environment):

```
REWARD_MERCURE=0 -> REWARD_MERCURE=1
```

The following variables have predefined values and those can be changed optionally:
```
MERCURE_SERVER_NAME=":80"
MERCURE_PUBLISHER_JWT_KEY="password"
MERCURE_PUBLISHER_JWT_ALG="HS256"
MERCURE_SUBSCRIBER_JWT_KEY="password"
MERCURE_SUBSCRIBER_JWT_ALG="HS256"
MERCURE_EXTRA_DIRECTIVES=""
```
