# Gogatekeeper Operator

This operator is built on the [Gatekeeper](https://github.com/gogatekeeper/gatekeeper) project, which is a proxy that
can be added to an application as a sidecar and provides the application with an [OIDC](https://openid.net/connect/)
authentication/authorization layer.

This operator is intended to accomplish two tasks:

1) Give users a simple method of declaring that an application should be given a `gatekeeper` sidecar and defining the
application specific configurations.

2) Allow users to define default OIDC gateways and configuration to be used with a gatekeeper instance.


## Usage

Option names map exactly (or almost exactly) in a few cases to their gatekeeper equivalents.
The best way to see all gatekeeper options is to run:

```bash
podman run gogatekeeper/gatekeeper:1.3.4 --help
```

### CRD

```yaml
apiVersion: gatekeeper.theendbeta.me/v1alpha1
kind: Gogatekeeper
metadata:
  name: gatekeeper-test
spec:
  # (required) OIDC discovery url for your provider
  oidcurl: http://127.0.0.1:5556/dex
  # (optional) default configuration to apply to all instances using this resource
  defaultconfig: |-
    upstream-url:          http://127.0.0.1:80
    listen:                :3000
    enable-refresh-tokens: true
    secure-cookie:         false
```

### Annotations

The required annotations must be on the `Pod` template, not the top-level `Deployment`, as the webhook currently works
at the `Pod` level.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-gk
  labels:
    app: nginx-gk-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-gk-test
  template:
    metadata:
      labels:
        app: nginx-gk-test
      annotations:
        # The name of the gatekeeper.theendbeta.me/v1alpha1 resource to use
        gatekeeper.gogatekeeper: gatekeeper-test
        # OIDC client-id for the application (`--client-id` option)
        gatekeeper.gogatekeeper/client-id: "example-app"
        # OIDC client-secret for the application (`--client-secret` option)
        gatekeeper.gogatekeeper/client-secret: ZXhhbXBsZS1hcHAtc2VjcmV0
        # session state encryption key (`--encryption-key` option)
        gatekeeper.gogatekeeper/encryption-key: VtMkufTWXWA5x83C
        # redirect URL for application (post login) (`--redirection-url` option)
        gatekeeper.gogatekeeper/redirection-url: "http://10.176.128.136:30001"
    spec:
      containers:
      - name: nginx
        image: nginx:stable
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-gk
  labels:
    app: nginx-gk-test
spec:
  type: NodePort
  ports:
  # The port should be the one that `gatekeeper` is listening on, not the upstream service
  - port: 3000
    nodePort: 30001
    protocol: TCP
    name: http
  selector:
    app: nginx-gk-test
```


## Install

### cert-manager

As this operator utilizes a webhook, it currently requires [cert-manager](cert-manager.io) to be installed on the
cluster for certificate generation.
You can follow their [installation guide](https://cert-manager.io/docs/installation/) for how to do this - no special
parameters are required.


### gogatekeeper-operator

**Note:** If you do not want to build/push your own copy of the operator image, you may immediately jump to
[Execute](#execute) after cloning the repository.

```bash
git clone https://github.com/theEndBeta/gogatekeeper-operator
cd gogatekeeper-operator

# If you want to build your own image locally
make docker-build
# -or-
make podman-build

# If you want to push said image to a repository
make docker-push
# -or-
make podman-push
```

## Execute

With your chosen `kubecontext` set as the current default, run the following to deploy the CRD(s) and controller to your
cluster.

```bash
make deploy
```


## Test

### Files

There are three example yaml files available in the `testfiles` directory.

* `gatekeeper.yaml`
  * an example gatekeeper CRD to deploy to the cluster.
  Defines an OIDC discovery URL for this configuration
* `nginx.yaml`
  * A Deployment for nginx *without* the required annotations for testing the negative case.
* `nginx-gatekeeper.yaml`
  * A Deployment for nginx *with* the required annotations for `gatekeeper` container injection.

### Testing Steps

0) Get access to an OIDC provider.

    If you don't already have access to an OIDC provider (or don't want to deal with creating a new client), you can run
    a local provider using [Dex](https://dexidp.io/docs/getting-started/), which has a convienent example configuration
    that requires only minor editing.
    (Add `<your k8s IP>:30001/oauth/callback` to `staticClients[0].redirectURIs` in `config-dev.yaml`)

    Gatekeeper was originally built to work with [Keycloak](https://www.keycloak.org), but that's a bit to heavy for a
    quick test.


1) Edit and apply a gatekeeper CRD

    You will first want to edit the `oidcurl` field on `testfiles/gatekeeper.yaml` to specify your OIDC providers
    discovery URL.

    Then run:

    ```bash
    kubectl apply -f testfiles/gatekeeper.yaml
    ```

    This will add the `gatekeeper-test` CRD and ConfigMap to the `default` namespace.

2) Apply the standalone nginx deployment

    `kubectl apply -f testfiles/nginx.yaml`

    This will create an solo nginx deployment in the `default` namespace listening on a `NodePort` of `30002`.
    Visiting `<your instance ip>:30002` should send you directly to the default nginx welcome page.

3) Edit and apply the nginx with gatekeeper deployment

    You will first want to edit the `gogatekeeper.gatekeeper` annotations on `testfiles/nginx-gatekeeper.yaml` to
    specify the following as defined by your OIDC provider:
    * `client-id`
    * `client-secret`
    * `redirection-url`

    Then run:

    `kubectl apply -f testfiles/nginx-gatekeeper.yaml`

    This will create a gatekeeper annotated nginx deployment in the `default` namespace listening on a `NodePort` of
    `30001`.
    This pod will have two containers - the nginx container and a gatekeeper container
    Visiting `<your instance ip>:30001` should redirect you to your OIDC provider's login page, where you can log in
    with your credentials and be redirected to the upstream nginx instance welcome page.
    (If you are using the Dex example deployment, the static example email/password is defined in `config-dev.yaml`).


## TODOs

* Add ability to specify `client-id`/`client-secret` as a k8s `Secret`
* Add ability to specify additional configuration fields in the gatekeeper CRD for defining the default gatekeeper
  configuration.
* Add update monitoring/handling to gatekeeper CRD admission webhook.
* Add validation webhook to gatekeeper CRD.
* Automated tests
