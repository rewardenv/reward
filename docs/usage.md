## Usage

### Common Commands

Run only the `db` container

``` bash
$ reward env up -- db
```

Launch a shell session within the project environment's `php-fpm` container:

``` bash
$ reward shell
```

Start a stopped environment:

``` bash
$ reward env start
```

Stop a running environment:

``` bash
$ reward env stop
```

Remove the environment and volumes completely:

``` bash
$ reward env down -v
```

Import a database:

``` bash
# for plain SQL database dump you can simply use os' stdin
$ reward db import < /path/to/dump.sql

# for compressed database dump
$ gunzip /path/to/dump.sql.gz -c | reward db import
```

Monitor database processlist:

``` bash
$ watch -n 3 "reward db connect -A -e 'show processlist'"
```

Tail environment nginx and php logs:

``` bash
$ reward env logs --tail 0 -f nginx php-fpm php-debug
```

Tail the varnish activity log:

``` bash
$ reward env exec -T varnish varnishlog
```

Clean varnish cache:

``` bash
$ reward env exec varnish varnishadm 'ban req.url ~ .'

# or you can use this, these commands are identical:
$ reward shell --container varnish varnishadm 'ban req.url ~ .'
```

Connect to redis:

``` bash
$ reward env exec redis redis-cli
```

Flush redis completely:

``` bash
$ reward env exec -T redis redis-cli flushall
```

### Further Information

You can call `--help` for any of reward's commands. For example `reward --help` or `reward env --help` for more
details and useful command information.
