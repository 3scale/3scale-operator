# Setup Redis Sentinel for TLS communication on OpenShift

## DRAFT, WIP

 ## Overview
 **Sentinel** is responsible for monitoring the Redis instance and handling failover in case the primary Redis instance goes down. You can scale the number of Sentinel nodes and configure your application to connect to Redis using the Sentinel service for high availability.

If both the **Redis server** (redis) and **Sentinel** (redis-sentinel) are using **TLS** for communication, you'll need to configure both the Redis server and the Redis Sentinel to use **SSL/TLS encryption** for secure connections. This involves setting up certificates and keys for both the Redis server and the Redis Sentinel, and then ensuring the communication between them and the clients is secured via TLS.

Here’s how to do this step-by-step in the OpenShift environment.
The Guide includes not only Sentinel TLS configuration, but Redis server also, to be useful for Testing. It will also include failover scenarion verification.

##  Redis Sentinel environment setup

### 1. **Prepare TLS Certificates**
Before you can configure Redis and Redis Sentinel to use TLS, you'll need to create and configure certificates for the Redis server, Redis Sentinel, and optionally for the client.

**Note**: To correctly generate the Redis server and Redis Sentinel certificates, you need to use their services IPs as the Common Names (CN) in the certificates. If you haven’t set up the services yet, please refer to **Section 4** to create the Redis server and Redis Sentinel services, before continue to next item. Once the services are running, use their respective IPs as the CN when generating the certificates.

#### a. Generate SSL Certificates (Self-signed or CA-signed)
You can create self-signed certificates or use certificates signed by a Certificate Authority (CA). Below is an example of generating self-signed certificates:
- Generate the CA private key and certificate
- Generate Redis server private key, CSR (Certificate Signing Request) and certificate
- Generate Redis Sentinel private key, CSR and certificate
- Generate Redis Client private key, CSR and certificate

Folloing example contains services IPs, please change it to yours.

```bash
openssl genpkey -algorithm RSA -out ca.key -pkeyopt rsa_keygen_bits:2048
openssl req -x509 -new -key ca.key -out ca.crt -days 3650 -subj "/CN=Redis CA"

openssl genpkey -algorithm RSA -out redis-server.key -pkeyopt rsa_keygen_bits:2048
openssl req -new -key redis-server.key -out redis-server.csr -subj "/CN=172.30.56.166"
openssl x509 -req -in redis-server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out redis-server.crt -days 365

openssl genpkey -algorithm RSA -out redis-sentinel.key -pkeyopt rsa_keygen_bits:2048
openssl req -new -key redis-sentinel.key -out redis-sentinel.csr -subj "/CN=172.30.217.195"
openssl x509 -req -in redis-sentinel.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out redis-sentinel.crt -days 365

openssl genpkey -algorithm RSA -out redis-client.key
openssl req -new -key redis-client.key -out redis-client.csr -subj "/CN=redis-client.example.com"
openssl x509 -req -in redis-client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out redis-client.crt -days 365
```

This will generate the following certificates and keys:
- `ca.crt`: The CA certificate
- `redis-server.crt` and `redis-server.key`: The Redis server’s certificate and key
- `redis-sentinel.crt` and `redis-sentinel.key`: The Redis Sentinel’s certificate and key
- `redis-client.crt` and `redis-client.key`: The Redis client's certificate and key

#### b. **Create a Secret in OpenShift**
You can store the certificates and keys in OpenShift as secrets for easy access.

```bash
oc create secret generic redis-tls-secret \
  --from-file=redis-server.crt \
  --from-file=redis-server.key \
  --from-file=ca.crt


oc create secret generic redis-sentinel-tls-secret \
  --from-file=redis-sentinel.crt \
  --from-file=redis-sentinel.key \
  --from-file=ca.crt
```

These secrets contain the necessary certificates for Redis and Redis Sentinel.

---

### 2. **Configure Redis Server for TLS**
You need to configure your Redis server to accept TLS connections by editing the `redis.conf` file. Below is an example configuration snippet:

```bash
# redis.conf

# Enable TLS
tls-port 6379
tls-cert-file /etc/redis/tls/redis-server.crt
tls-key-file /etc/redis/tls/redis-server.key
tls-ca-cert-file /etc/redis/tls/ca.crt

# Require clients to authenticate via TLS
tls-auth-clients no

# Enable Redis to listen for connections using TLS
port 0  # Disable non-TLS port (optional)
```
```yaml
apiVersion: v1
data:
  redis.conf: |+
    # redis.conf
    bind 0.0.0.0
    protected-mode no
    port 6379
    tls-port 6380
    tls-cert-file /etc/redis/certs/redis-server.crt
    tls-key-file /etc/redis/certs/redis-server.key
    tls-ca-cert-file /etc/redis/certs/ca.crt
    tls-auth-clients yes
    stop-writes-on-bgsave-error no
    save ""

kind: ConfigMap
metadata:
  name: redis-config-redis
```



Now, in your **OpenShift Redis deployment**, mount the secrets as volumes and ensure the correct paths are used.

```yaml
# redis-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: quay.io/fedora/redis-6
        ports:
        - containerPort: 6379
        volumeMounts:
        - name: redis-config-volume
          mountPath: /etc/redis/redis.conf
          subPath: redis.conf
        - name: redis-tls-volume
          mountPath: /etc/redis/certs
          readOnly: true
        command: ["/bin/sh", "-c", "redis-server /etc/redis/redis.conf"]
      volumes:
      - name: redis-config-volume
        configMap:
          name: redis-config-redis
      - name: redis-tls-volume
        secret:
          secretName: redis-tls-secret

```

Then, apply this configuration to OpenShift:

```bash
oc apply -f redis-deployment.yaml
```

---

### 3. **Configure Redis Sentinel for TLS**
Redis Sentinel also needs to be configured to support TLS communication.

- ConfigMap: `redis-sentinel-config.yaml`
  
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-sentinel-config
data:
  sentinel.conf: |
    # Basic Redis Sentinel configuration
    # Non-TLS port for Redis Sentinel
    port 26379                          # Default non-TLS port for Sentinel

    # TLS port for Redis Sentinel
    tls-port 26380                       # TLS port for Redis Sentinel
    tls-cert-file /etc/ssl/certs/redis-sentinel.crt  # Path to the Redis Sentinel certificate
    tls-key-file /etc/ssl/private/redis-sentinel.key  # Path to the Redis Sentinel private key
    tls-ca-cert-file /etc/ssl/certs/ca.crt  # Path to the CA certificate (for both client and server verification)
    tls-auth-clients yes                 # Require clients to present certificates (mutual TLS)

    # Optional: Configure a set of supported ciphers
    tls-ciphersuites TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256
    tls-protocols TLSv1.2 TLSv1.3          # Use only TLSv1.2 and TLSv1.3

    # Monitoring configuration
    sentinel monitor mymaster redis 6380 2  # Monitor the Redis master at 'redis' on port 6379
    sentinel down-after-milliseconds mymaster 5000  # If Redis is down for 5 seconds, trigger failover
    sentinel failover-timeout mymaster 10000       # Timeout for failover
    sentinel parallel-syncs mymaster 1             # Number of replicas to synchronize during failover

    # Optional: If you're using a password for Sentinel authentication
    requirepass <your_redis_password>             # Optional password for Redis Sentinel

    # Optional: Other security settings
    # You can configure additional parameters, such as monitoring a specific Redis replica, setting timeouts, etc.
    timeout 300                                   # Optional: Timeout for idle clients

```

Again, you’ll mount the Sentinel certificates and keys as volumes, similar to what you did with Redis.

```yaml
# redis-sentinel-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-sentinel
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis-sentinel
  template:
    metadata:
      labels:
        app: redis-sentinel
    spec:
      containers:
        - name: redis-sentinel
          image: redis:latest
          ports:
            - containerPort: 26379   # Non-TLS port
            - containerPort: 26380   # TLS port
          volumeMounts:
            - name: redis-sentinel-config-volume
              mountPath: /etc/redis
              subPath: sentinel.conf
            - name: redis-tls-secret
              mountPath: /etc/ssl/certs
              readOnly: true   # Ensure certificates are read-only
      volumes:
        - name: redis-sentinel-config-volume
          configMap:
            name: redis-sentinel-config  # The ConfigMap containing the sentinel.conf file
        - name: redis-tls-secret
          secret:
            secretName: redis-sentinel-tls-secret  # The secret holding your TLS certificates (redis-sentinel.crt, redis-sentinel.key, ca.crt)
```

Apply the Redis Sentinel deployment:

```bash
oc apply -f redis-sentinel-deployment.yaml
```

---

### 4. Update Redis Service and Sentinel Service
You should ensure that the Redis and Redis Sentinel services are properly configured to use the TLS ports.

#### Redis Service (TLS port 6379)
```yaml
# redis-service.yaml

apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  ports:
    - port: 6379         # Non-TLS (unencrypted) port
      targetPort: 6379
      name: redis
    - port: 6380         # TLS port
      targetPort: 6380
      name: redis-tls
  selector:
    app: redis
  type: NodePort 

```

#### Redis Sentinel Service
```yaml
# redis-sentinel-service.yaml

apiVersion: v1
kind: Service
metadata:
  name: redis-sentinel
spec:
  selector:
    app: redis-sentinel
  ports:
    - name: redis       # Name of the port for non-TLS traffic
      protocol: TCP
      port: 26379        # The external service port for non-TLS connections
      targetPort: 26379   # The internal port for non-TLS traffic
    - name: redis-tls    # Name of the port for TLS traffic
      protocol: TCP
      port: 26380        # The external service port for TLS connections
      targetPort: 26380   # The internal port for TLS traffic

```

Apply both services:

```bash
oc apply -f redis-service.yaml
oc apply -f redis-sentinel-service.yaml
```

- Note that services IPs will be used to create certificates.

---

### 5. **Connect to Redis with TLS**

To connect to the Redis server and Redis Sentinel with TLS using `redis-cli`, you need to specify the appropriate certificates and TLS options.

#### **Redis CLI (from a client)**
```bash
# Connect to Redis server with TLS
redis-cli -h redis -p 6379 --tls --cert /path/to/client.crt --key /path/to/client.key --cacert /path/to/ca.crt

# Connect to Redis Sentinel with TLS
redis-cli -h redis-sentinel -p 26379 --tls --cert /path/to/client.crt --key /path/to/client.key --cacert /path/to/ca.crt SENTINEL get-master-addr-by-name mymaster
```

If you're inside an OpenShift pod and want to use `redis-cli` from within a container:

```bash
oc rsh <sentinel-pod-name>
redis-cli -h redis-sentinel -p 26379 --tls --cert /etc/redis/tls/client.crt --key /etc/redis/tls/client.key --cacert /etc/redis/tls/ca.crt SENTINEL get-master-addr-by-name mymaster
```

---

### Conclusion

By following these steps, you’ve configured Redis and Redis Sentinel to communicate over **TLS** in your OpenShift cluster. This ensures secure communication between the Redis server and Redis Sentinel, and between your clients and the Redis instances (see next section for Clients interation details).

Key steps include:
- Setting up certificates for Redis and Sentinel.
- Configuring Redis and Sentinel to use TLS with the proper configuration files (`redis.conf` and `sentinel.conf`).
- Using OpenShift secrets to securely store certificates and keys.
- Ensuring your Redis clients (either inside or outside OpenShift) are configured to communicate using TLS.  
  


## Client interaction with the Redis Sentinel and Redis server 

In a **Redis Sentinel** environment, the client application does **not** directly connect to the **Redis Sentinel** nodes. Instead, the client application connects to the **Redis server** (e.g., the `redis` server or the current master Redis instance) for actual data operations (reads and writes). However, the client needs to interact with the Redis Sentinel to determine which Redis server is currently the **master** in case of failover or discovery.

Here's a detailed explanation of how the client application interacts with the Redis Sentinel and Redis server:

### 1. **Role of Redis Sentinel**
Redis Sentinel is primarily responsible for:
- **Monitoring** the Redis servers (e.g., `redis`).
- **Failover management**: In case the master Redis server goes down, Sentinel promotes one of the replicas to become the new master.
- **Notification**: It informs clients about the current master server.

### 2. **How Clients Interact with Redis Sentinel and Redis Servers**

- **Initial Connection**: The client application does **not** connect directly to the Redis Sentinel nodes for data operations. Instead, it first connects to the **Redis Sentinel** to determine which Redis server is currently the **master**.
  
- **Getting the Current Master**: The client sends a `SENTINEL get-master-addr-by-name mymaster` request to the Sentinel to retrieve the address (IP and port) of the **current master** Redis instance. The Sentinel provides this information based on the monitoring data it has.
  
  - The client sends this query to **any Sentinel** node, and the Sentinel responds with the IP address and port of the current master.

- **Connecting to the Redis Master**: After obtaining the master information from the Sentinel, the client connects directly to the **Redis server** (the current master) for actual data operations (e.g., `SET`, `GET`, etc.).

- **Failover Handling**: If the client is connected to a master Redis instance and that master goes down, Redis Sentinel will promote one of the replica instances to be the new master. The client can then query Redis Sentinel again to get the updated address of the new master. This allows the client to continue its operations without manual intervention.


### 3. **How Failover Works with Sentinel**
In case of a failover (for example, when the current master goes down), Redis Sentinel promotes a replica to be the new master. The client is expected to:

1. **Detect Master Change**: The client may receive a notification or keep querying Redis Sentinel to get the address of the current master.
   
2. **Re-establish Connection**: Upon receiving the new master’s address from Sentinel, the client automatically switches to the new master and resumes data operations.

Redis Sentinel does not handle actual data requests. It only provides information about the master server and handles the failover process. The client needs to be able to handle the transition to a new master automatically.

---

### 4. **Summary**
- **Client connects to Redis Sentinel** to discover the current master.
- **Client connects to the current master Redis server** for data operations (reads/writes).
- **In case of failover**, the client queries Redis Sentinel again to get the new master’s address and re-establish the connection.

So, **client applications never directly communicate with the Redis Sentinel** for data operations. They always connect to the Redis server (master) that Redis Sentinel has informed them about. The role of Redis Sentinel is to monitor Redis instances, handle failover, and provide the current master’s address to the client.

# **WIP**