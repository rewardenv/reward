## Nginx Configuration

It is possible to use custom nginx configurations in various ways. When you run `reward env up` it will
map `./.reward/nginx` directory to the container under `/etc/nginx/snippets` directory.

### Nginx Application template

You can override Nginx HTTP and Server blocks as well.

Nginx default.conf looks like this by default:
```
include /etc/nginx/snippets/http-*.conf;

server {
    listen 80;

    root ${NGINX_ROOT}${NGINX_PUBLIC};
    set $MAGE_ROOT ${NGINX_ROOT};

    index index.html index.php;
    autoindex off;
    charset UTF-8;

    include /etc/nginx/snippets/server-*.conf;
    include /etc/nginx/available.d/${NGINX_TEMPLATE};
}
```

Using this it is possible to inject configurations in 2 places.

``` note::
    If you'd like to create mappings (which have to be under nginx HTTP block) you can define templates in the
    `./.reward/nginx` directory using `http-*.conf` pattern.
    For example:

    .. code::

        ./.reward/nginx/http-example-mappings.conf
```

``` note::
    If you'd like to create redirects (which have to be under nginx servers block) you can define templates in the
    `./.reward/nginx` directory using `server-*.conf` pattern.
    For example:

    .. code::

        ./.reward/nginx/server-example-redirects.conf
```

#### Example: Create a rewrite from example.test/blog/* to blog.example.test/*

Create a file under the `./.reward/nginx` directory called server-redirect.conf.

``` bash
$ echo -e 'location ^~ /blog/ {
    rewrite ^/blog/(.*) https://blog.$http_host/$1 permanent;
  }' > ./.reward/nginx/server-redirect.conf

$ reward env restart -- nginx

# Test using curl
$ curl -IL https://example.test/blog/test-url
HTTP/2 301
content-type: text/html
date: Fri, 29 Jan 2021 14:12:22 GMT
location: https://blog.example.test/test-url
server: nginx/1.16.1
content-length: 169

```

