## SPX Support

To enable SPX support, you need to set the `REWARD_SPX` environment variable in the `.env` file to `true`.

Then run `reward env up` to apply the changes.

In this case a new container will be created, called `php-spx`, which will have the SPX extension installed.

All requests that contain one of the following will be forwarded to the `php-spx` container:

- `SPX_ENABLED` cookie
- `SPX_KEY` cookie
- `SPX_ENABLED` argument in the URL (e.g. `https://m2.test/?SPX_ENABLED=1`)
- `SPX_KEY` argument (e.g. `?SPX_KEY=dev`)
- `SPX_UI_URI` argument (e.g. `?SPX_UI_URI=/`)

``` warning::
    If you **open the SPX UI**, the **SPX cookies will be automatically set** in your browser. If you want to disable 
    sending all requests to the SPX container, you need to clear the cookies.
```

In similar fashion to the `reward shell` command there is also an spx command to launch into an SPX enabled
container shell for debugging CLI workflows:

```
reward spx
```

In this container if you run any commands those will be executed with the SPX profiling enabled.

### SPX UI

To access the SPX UI, you need to navigate to the environment URL with the `SPX_UI_URI` argument set to `/`.

E.g.: `https://m2.test/?SPX_KEY=dev&SPX_UI_URI=/`

If you run `reward info` you will see the SPX UI URL in the output.

### SPX Key

The `SPX_HTTP_KEY` environnment variable can be set to a value of your choice.
This key will be used to authenticate requests to the SPX container.

It defaults to be empty.

If you configure the `SPX_HTTP_KEY` environment variable, you will need to pass the key in the URL to access the SPX UI.

E.g.: `https://m2.test/?SPX_KEY=dev&SPX_UI_URI=/`

But `reward info` will show you the correct URL to access the SPX UI.

### Configuration

Reference documentation for the SPX variables are
available [here](https://github.com/NoiseByNorthwest/php-spx?tab=readme-ov-file#advanced-usage).

The following variables can be configured using environment variables:

- `SPX_DEBUG` - default: `1`
- `SPX_HTTP_ENABLED` - default: `1`
- `SPX_HTTP_KEY` - default: `""`
- `SPX_HTTP_IP_WHITELIST` - default: `*`
- `SPX_HTTP_PROFILING_ENABLED` - default: `1`
- `SPX_HTTP_PROFILING_AUTO_START` - default: `1`
- `SPX_HTTP_TRUSTED_PROXIES` - default: `*`
