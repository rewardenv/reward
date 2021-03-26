## Usage

### Common Commands

Run only the `db` container

``` bash
$ reward env up -- db
```

Launch a shell session within the project environment's `php-fpm` container:

``` bash
$ reward shell

# or launch an `sh` shell in the nginx container
$ reward shell sh --container nginx
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

Run complex MySQL queries directly using `reward db connect`:

``` bash
# Note: to pass arguments use double dash to terminate Reward's argument parsing and escape the special characters [;'"]*
# Run inline query:
$ reward db connect -- -e \"SELECT table_name FROM information_schema.tables WHERE table_schema=\'magento\' ORDER BY table_name LIMIT 5\;\"

# Run query passing a bash variable (note the escaped quote):
$ MYSQL_CMD="\"SELECT table_name FROM information_schema.tables WHERE table_schema='magento' ORDER BY table_name LIMIT 5;\""

$ reward db connect -- -e $MYSQL_CMD

# Run multiple queries/commands using heredoc:
$ MYSQL_CMD=$(cat <<"EOF"
"SELECT table_name FROM information_schema.tables WHERE table_schema='magento' ORDER BY table_name LIMIT 5;
QUERY2;
QUERY3;"
EOF
)

$ reward db connect -- -e $MYSQL_CMD
```

Dump database:

```
$ reward db dump | gzip -c > /path/to/db-dump.sql.gz
```

Monitor database processlist:

``` bash
$ watch -n 3 "reward db connect -- -e \'show processlist\'"
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
