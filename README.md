# Traefik Plugin - Fast Redirect

`Fast Redirect` is a Traefik plugin to redirect a list with status code.

Based on :
- [traefik-plugin-redirect by evolves-fr](https://github.com/evolves-fr/traefik-plugin-redirect)
- [Traefik documentation](https://doc.traefik.io/traefik-pilot/plugins/overview/)
- [Traefik plugin example](https://github.com/traefik/plugindemo)
- [Traefik internal redirect plugin](https://github.com/traefik/traefik/blob/master/pkg/middlewares/redirect/redirect.go)

## Installation

Into Traefik static configuration

### TOML
```toml
[entryPoints]
  [entryPoints.web]
    address = ":80"

[pilot]
  token = "xxxxxxxxx"

[experimental.plugins]
  [experimental.plugins.traefik-plugin-redirect]
    moduleName = "github.com/epicagency/traefik-plugin-redirect"
    version = "v1.0.0"
```

### YAML
```yaml
entryPoints:
  web:
    address: :80

pilot:
    token: xxxxxxxxx

experimental:
  plugins:
    traefik-plugin-redirect:
      moduleName: "github.com/epicagency/traefik-plugin-redirect"
      version: "v1.0.0"
```

### CLI
```shell
--entryPoints.web.address=:80
--pilot.token=xxxxxxxxx
--experimental.plugins.traefik-plugin-redirect.modulename=github.com/epicagency/traefik-plugin-redirect
--experimental.plugins.traefik-plugin-redirect.version=v1.0.0
```

## Configuration

Into Traefik dynamic configuration

### Docker
```yaml
labels:
  - "traefik.http.middlewares.my-redirect.plugin.redirect.redirects[0]=/301:/moved-permanently:301"
  - "traefik.http.middlewares.my-redirect.plugin.redirect.redirects[1]=/302:/implicit-temporary-redirect"
  - "traefik.http.middlewares.my-redirect.plugin.redirect.redirects[2]=/not-found::404"
```

### Kubernetes
```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: my-redirect
spec:
  plugin:
    traefik-plugin-redirect:
      redirects:
      - /301:/moved-permanently:301
      - /302:/implicit-temporary-redirect
      - /not-found::404
```

### TOML
```toml
[http]
  [http.middlewares]
    [http.middlewares.my-redirect]
      [http.middlewares.my-redirect.plugin]
        [http.middlewares.my-redirect.plugin.traefik-plugin-redirect]
        redirects =[
        "/301:/moved-permanently:301",
        "/302:/implicit-temporary-redirect",
        "/not-found::404"
        ]
```

### YAML
```yaml
http:
  middlewares:
    my-redirect:
      plugin:
        traefik-plugin-redirect:
          redirects:
          - /301:/moved-permanently:301
          - /302:/implicit-temporary-redirect
          - /not-found::404
```
