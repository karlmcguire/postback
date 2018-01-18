# Development

## Server Configuration (45 minutes)

After reading the project spec, it looks like I just need PHP, Redis, and Go on the server. I'll install everything on the box now and consider making separate containers later.

### Ingestion Agent

I chose `nginx` because I prefer it and I assume it's being used in production at Kochava. The rest are just standard install commands.

```
$ apt-get install nginx
$ apt-get install php-fpm
$ apt-get install redis-server
$ apt-get install php-redis
```

Next I need to configure `/etc/nginx/sites-available/default`. `INGEST_IP` is the address of the box provided to me. I uncommented the lines needed for `php-fpm` to handle PHP files.

```nginx
server_name [INGEST_IP];

location ~ \.php$ {
	include snippets/fastcgi-php.conf;
	fastcgi_pass unix:/run/php/php7.0-fpm.sock;
}
```

I also need to configure `/etc/redis/redis.conf`. I changed the default port to `INGEST_PORT` and added authentication with `INDEX_PASS` (which is a very long hash).

```conf
port [INGEST_PORT]
requirepass [INGEST_PASS]
```

Now that I've configured `nginx`, `php-fpm`, and `redis`, I can restart each service and enable them on boot. 

```
$ systemctl restart nginx
$ systemctl enable nginx
```

```
$ systemctl restart php7.0-fpm
$ systemctl enable php7.0-fpm
```

```
$ systemctl restart redis-server.service
$ systemctl enable redis-server.service
```

## Development

I find that I'm most effective when I have a clear mental model of the application architecture. So, I created a diagram based on the project spec to assist me in keeping things organized.

<p align="center">
	<img src="img/1.svg" />
</p>

I'll worry about turning each component into it's own container later. For now I'll just write the code. I'll start with the Ingestion Agent. 

### Ingestion Agent
















