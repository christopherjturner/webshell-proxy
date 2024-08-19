# WebShell Proxy

## Usage

`./webshell-proxy -p 8085`

or:

```
export PORT=8085
./webshell-proxy
```

## Register a route

Temp api


Register a new route
```
curl https://address-of-proxy/register?id=<token>&target=http://addr-of-webshell-instance:8085"
```


Get a list of all routes registered
```
curl https://address-of-proxy/routes
```

