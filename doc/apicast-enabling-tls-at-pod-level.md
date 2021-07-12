## APIcast: Enabling TLS at pod level

3scale managed APIcast pods can be deployed to serve HTTP endpoints over TLS. Each APIcast instance, i.e. `production` and `staging`, can be individually configured to serve, or not, HTTP over TLS.

Enable TLS at APIcast pod level setting either `httpsPort` or `httpsCertificateSecretRef` fields or both.

Steps to enable TLS at pod level:

1.- Generate self signed certificates for your DOMAIN **[Optional]**

```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout server.key -out server.crt
```

Fill out the prompts appropriately. The most important line is the one that requests the Common Name (e.g. server FQDN or YOUR name). You need to enter the domain name associated with your server or, more likely, your server’s public IP address.

2.- Create the certificate secret

```
kubectl create secret tls mycertsecret --cert=server.crt --key=server.key
```

3.- Reference the certificate secret in APIManager CR

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager-apicast-custom-environment
spec:
  wildcardDomain: <desired-domain>
  apicast:
    productionSpec:
      httpsPort: 8443
      httpsCertificateSecretRef:
        name: mycertsecret
```

**NOTE 1**: If `httpsPort` is set and `httpsCertificateSecretRef` is not set, APIcast will use a default certificate bundled in the image.

**NOTE 2**: If `httpsCertificateSecretRef` is set and `httpsPort` is not set, APIcast will enable TLS at port number **8443**.

**NOTE 3**: The example above deploys *production* APIcast to serve TLS. The example for the *staging* APIcast would as follows:

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager-apicast-custom-environment
spec:
  wildcardDomain: <desired-domain>
  apicast:
    stagingSpec:
      httpsPort: 8443
      httpsCertificateSecretRef:
        name: mycertsecret
```

See [APIManager CRD reference](apimanager-reference.md) for all available options.

The TLS port can be accessed using apicast service's named port `httpsproxy`. You can check using `oc port-forward` command.

Open a terminal and run the port forwarding command for `httpsproxy` named port.

```
$ oc port-forward service/apicast-production httpsproxy
Forwarding from 127.0.0.1:8443 -> 8443
Forwarding from [::1]:8443 -> 8443
```

In other terminal, download used certificate.

```
$ echo quit | openssl s_client -showcerts -connect 127.0.0.1:8443 2>/dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p'
```

The downloaded certificate should match provided certificate.
