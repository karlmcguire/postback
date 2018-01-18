# Development

## Server Configuration (30 minutes)

After reading the project spec, it looks like I just need PHP, Redis, and Go on the server. I'll install everything on the box now and consider making separate containers later.

### Ingestion Agent

I chose `nginx` because I prefer it and I assume it's being used in production at Kochava. The rest are just standard install commands.

```
$ apt-get install nginx
```

```
$ apt-get install php-fpm
```

```
$ apt-get install redis-server
```

```
$ apt-get install php-redis
```

Next I need to configure `/etc/nginx/sites-available/default`. `INGEST_IP` is the address of the box provided to me. I uncommented the lines needed for `php-fpm` to handle PHP files.

```nginx
server_name [INGEST_IP];

location ~ \.php$ {
	include snippets/fastcgi-php.conf;
	fastcgi_pass unix:/run/php/php7.0-fpm.sock;
}

location ~ /\.ht {
	deny all;
}
```









