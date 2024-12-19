# Configuring Redis Sentinel for TLS on OpenShift

This document provides an overview of Redis Sentinel with TLS, detailing the installation and configuration process (without using a Helm Chart).

- [Configuring Redis Sentinel for TLS on OpenShift](#configuring-redis-sentinel-for-tls-on-openshift)
  - [Overview](#overview)
    - [Redis and Redis-Sentinel (Sentinel): Same Image, Different Configuration](#redis-and-redis-sentinel-sentinel-same-image-different-configuration)
      - [Image Relationship](#image-relationship)
      - [Key Points](#key-points)
  - [Sentinel environment setup](#sentinel-environment-setup)
    - [Prepare TLS Certificates](#prepare-tls-certificates)
      - [Create Redis Service](#create-redis-service)
      - [Generate SSL Certificates (Self-signed or CA-signed)](#generate-ssl-certificates-self-signed-or-ca-signed)
      - [Create Redis and Sentinel secrets](#create-redis-and-sentinel-secrets)
    - [Install Redis Server with TLS support](#install-redis-server-with-tls-support)
      - [Redis ConfigMap:](#redis-configmap)
      - [Redis StatefulSet](#redis-statefulset)
    - [Install Sentinel with TLS support](#install-sentinel-with-tls-support)
      - [Sentinel Service](#sentinel-service)
      - [Sentinel ConfigMap](#sentinel-configmap)
      - [Sentinel StatefulSet](#sentinel-statefulset)
    - [Installation Results](#installation-results)
    - [Summary](#summary)
  - [The Process of Client interaction with the Sentinel and Redis server](#the-process-of-client-interaction-with-the-sentinel-and-redis-server)
    - [Role of Sentinel](#role-of-sentinel)
    - [How Clients Interact with Sentinel and Redis Servers](#how-clients-interact-with-sentinel-and-redis-servers)
    - [How Failover Works with Sentinel](#how-failover-works-with-sentinel)
    - [Summary](#summary-1)
  - [Sentinel Test on EC2 (briefly)](#sentinel-test-on-ec2-briefly)
  - [Some useful commands](#some-useful-commands)


## Overview

### Redis and Redis-Sentinel (Sentinel): Same Image, Different Configuration

*We will use term **Sentinel** instead of full name **Sentinel** in most places in this doc.*

1. **Redis**:
   - Redis is an in-memory key-value store that you use for caching, session management, pub/sub messaging, etc.
   - The Redis service itself is a single standalone instance that serves as the database.

2. **Sentinel**:
   - Sentinel is a system designed to manage Redis instances, particularly in high availability (HA) and failover scenarios. 
It is not a separate Redis product but rather a set of services that provide monitoring, failover, and notification for Redis.
   - Sentinel can be considered as an extension of Redis that works alongside one or more Redis instances to monitor their health and ensure redundancy by promoting a new master if the current master fails.

#### Image Relationship

- **Same Image**: Redis and Sentinel are **the same software**. The Sentinel functionality is bundled within the Redis image itself. The  Sentinel process is included, but **it is not enabled by default**. You have to explicitly configure and run the Sentinel process separately, usually on its own set of nodes, alongside your Redis instances.

- **Configuration and Launching**:
  - To run Sentinel, you need to configure a separate Sentinel configuration file (`sentinel.conf`) and launch it with the `redis-server` binary by passing the `--sentinel` flag (for example: `redis-server /etc/redis/sentinel.conf --sentinel`).
  - Sentinel uses the same Redis image, but you run it with a different configuration to enable Sentinel's functionality.

#### Key Points
- **Same image**: Redis and Sentinel share the same Redis image.
- **Different configurations**: Redis runs as a normal database, while Sentinel is configured and run as a process to monitor and manage Redis instances.
- **Redis and Sentinel are typically run as separate processes**: You run one or more Redis instances and multiple Sentinel instances (usually 3 or more) for high availability.

 **Sentinel** is responsible for monitoring the Redis instance and handling failover in case the primary Redis instance goes down. You can scale the number of Sentinel nodes and configure your application to connect to Redis using the Sentinel service for high availability.

If both the **Redis server** (redis) and **Sentinel** (redis-sentinel) are using **TLS** for communication, you'll need to configure both the Redis server and the Sentinel to use **SSL/TLS encryption** for secure connections. This involves setting up certificates and keys for both the Redis server and the Sentinel, and then ensuring the communication between them and the clients is secured via TLS.

Here’s how to do this step-by-step in the OpenShift environment.
The Guide includes not only Sentinel TLS configuration, but Redis server also, to be useful for Testing. 
It will also include failover scenarion verification.

##  Sentinel environment setup

### Prepare TLS Certificates
Before you can configure Redis and Sentinel to use TLS, you'll need to create and configure certificates for the Redis server, Sentinel, and optionally for the client.

**Note**: To correctly generate the Redis server and Sentinel certificates, you need to use their services IPs as the Common Names (CN) in the certificates. If you haven’t set up the services yet, please refer to **Section 4** to create the Redis server and Sentinel services, before continue to next item. Once the services are running, use their respective IPs as the CN when generating the certificates.

#### Create Redis Service 
- Service is created with ClusterIP type (default, not defined) - for Internal cluster usage and for HA.

```yaml
cat << EOF | oc create -f -
apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  selector:
    app: redis
  ports:
    - port: 6379         # Non-TLS (unencrypted) port
      targetPort: 6379
      name: redis
    - port: 6380         # TLS port
      targetPort: 6380
      name: redis-tls
EOF
```

- Expected service example:

```bash
$oc get svc
NAME    TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
redis   ClusterIP   172.30.115.228   <none>        6379/TCP,6380/TCP   6s
```

#### Generate SSL Certificates (Self-signed or CA-signed)
You can create self-signed certificates or use certificates signed by a Certificate Authority (CA). Below is an example of generating self-signed certificates:
- Generate the CA private key and certificate
- Generate Redis server private key, CSR (Certificate Signing Request) and certificate
- Generate Sentinel private key, CSR and certificate
- Generate Redis Client private key, CSR and certificate

Folloing example contains services IPs, please change it to yours.

```bash
openssl genpkey -algorithm RSA -out ca.key -pkeyopt rsa_keygen_bits:2048
openssl req -x509 -new -key ca.key -out ca.crt -days 3650 -subj "/CN=Redis CA"

openssl genpkey -algorithm RSA -out redis-server.key -pkeyopt rsa_keygen_bits:2048
openssl req -new -key redis-server.key -out redis-server.csr -subj "/CN=172.30.115.228"
openssl x509 -req -in redis-server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out redis-server.crt -days 365

openssl genpkey -algorithm RSA -out redis-sentinel.key -pkeyopt rsa_keygen_bits:2048
openssl req -new -key redis-sentinel.key -out redis-sentinel.csr -subj "/CN=redis-sentinel.3scale-test.svc.cluster.local"
openssl x509 -req -in redis-sentinel.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out redis-sentinel.crt -days 365

openssl genpkey -algorithm RSA -out redis-client.key
openssl req -new -key redis-client.key -out redis-client.csr -subj "/CN=redis-client.example.com"
openssl x509 -req -in redis-client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out redis-client.crt -days 365
```

This will generate the following certificates and keys:
- `ca.crt`: The CA certificate
- `redis-server.crt` and `redis-server.key`: The Redis server’s certificate and key
- `redis-sentinel.crt` and `redis-sentinel.key`: The Sentinel’s certificate and key
- `redis-client.crt` and `redis-client.key`: The Redis client's certificate and key

**Notes** We use DNS for CN for Sentinel because Sentinel service should be headless. 
- A headless service (clusterIP: None) is used because Kubernetes should not assign a single ClusterIP to the Sentinel service. 
Instead, Sentinel should access each individual Sentinel pod directly via DNS names like `redis-sentinel-0.redis-sentinel.svc.cluster.local` and `redis-sentinel-1.redis-sentinel.svc.cluster.local`.
- Headless service provides DNS records for each pod, allowing Sentinel to interact with the individual Sentinel instances, 
which is essential for its monitoring and failover mechanisms.


#### Create Redis and Sentinel secrets
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

These secrets will contain the necessary certificates for Redis and Sentinel.


### Install Redis Server with TLS support

#### Redis ConfigMap:

```yaml
cat << EOF | oc create -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config-redis  
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
    tls-auth-clients no
    stop-writes-on-bgsave-error no
    save ""
EOF
```

#### Redis StatefulSet

```yaml
cat << EOF | oc create -f -
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
spec:
  serviceName: "redis"
  replicas: 3
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
        #image: quay.io/fedora/redis-6
        image: redis:7
        ports:
          - containerPort: 6379
          - containerPort: 6380
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
EOF
```


### Install Sentinel with TLS support

#### Sentinel Service

```yaml
cat << EOF | oc create -f -
apiVersion: v1
kind: Service
metadata:
  name: redis-sentinel
spec:
  selector:
    app: redis-sentinel
  ports:
    - name: redis 
      protocol: TCP
      port: 26379 
      targetPort: 26379 
    - name: redis-tls
      protocol: TCP
      port: 26380 
      targetPort: 26380
  clusterIP: None  # This makes the service headless
EOF
```

#### Sentinel ConfigMap
  
```yaml
cat << EOF | oc create -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-sentinel-config 
data:
  sentinel.conf: |
    port 0
    tls-port 26380
    tls-cert-file /etc/redis/certs/redis-sentinel.crt
    tls-key-file /etc/redis/certs/redis-sentinel.key
    tls-ca-cert-file /etc/redis/certs/ca.crt
    tls-auth-clients yes 
    tls-replication yes

    #sentinel monitor mymaster redis.3scale-test.svc.cluster.local 6380 2
    sentinel monitor mymaster  172.30.115.228 6380 2
     
    sentinel down-after-milliseconds mymaster 5000
    sentinel failover-timeout mymaster 10000
    sentinel parallel-syncs mymaster 1
EOF
```
**Notes** 
- please make sure use correct Redis service IP in option (example): `sentinel monitor mymaster  172.30.115.228 6380 2`. **Better - to use DNS**. We used IP for testing.
- if `port 0` is set (as in this example) - the default Sentinel non-TLS port `26379`  will be desabled, and Sentinel will use only TLS communication. 
- If you will define both `port 26379` (default) and `tls-port 26380`   - Sentinel will use both TLS and non TLS communications. We found that in logs it looks like Sentinel has two replicas that connect to Redis, TLS and non-Tls (details are not provided here).


#### Sentinel StatefulSet

```yaml
cat << EOF | oc create -f -
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-sentinel
spec:
  serviceName: "redis-sentinel"
  replicas: 3
  selector:
    matchLabels:
      app: redis-sentinel
  template:
    metadata:
      labels:
        app: redis-sentinel
    spec:
      initContainers:
        - name: copy-sentinel-config
          image: quay.io/fedora/redis-6
          command: ["/bin/sh", "-c", "cp /etc/redis/sentinel.conf /mnt/config/sentinel.conf"]
          volumeMounts:
            - name: redis-config-volume
              mountPath: /etc/redis/sentinel.conf
              subPath: sentinel.conf  # Read-only mount from ConfigMap
            - name: sentinel-config-writable
              mountPath: /mnt/config  # Writable volume for copying the file
      containers:
        - name: redis-sentinel
          #image: quay.io/fedora/redis-6
          image: redis:7
          ports:
            - containerPort: 26379
            - containerPort: 26380
          volumeMounts:
            - name: sentinel-config-writable
              mountPath: /mnt/config  # Use writable volume for accessing sentinel.conf
            - name: redis-tls-volume
              mountPath: /etc/redis/certs
              readOnly: true
          # Command now points to the correct location in the writable volume
          command: ["/bin/sh", "-c", "redis-server /mnt/config/sentinel.conf --sentinel"]
      volumes:
        - name: redis-config-volume
          configMap:
            name: redis-sentinel-config
        - name: sentinel-config-writable
          emptyDir: {}  # Writable volume to store the sentinel.conf
        - name: redis-tls-volume
          secret:
            secretName: redis-sentinel-tls-secret
EOF
```

**Notes** redis-sentinel by design needs to have write access on its own configuration file otherwise it will exit if it cannot write to it. To address this issue - Init container was added that copies readonly `/etc/redis/sentinel.conf` file created from config map to new location  `/mnt/config/sentinel.conf `, and use new writable configuration file in main container: `command: ["/bin/sh", "-c", "redis-server /mnt/config/sentinel.conf --sentinel"]`

### Installation Results

- expected running pods:
```bash
$oc get pod
NAME               READY   STATUS    RESTARTS   AGE
redis-0            1/1     Running   0          15h
redis-1            1/1     Running   0          15h
redis-2            1/1     Running   0          15h
redis-sentinel-0   1/1     Running   0          15h
redis-sentinel-1   1/1     Running   0          15h
redis-sentinel-2   1/1     Running   0          15h
```
- Redis - Ready to accept connections tls
```bash
$oc logs redis-0
7:C 11 Dec 2024 16:05:46.035 * oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
7:C 11 Dec 2024 16:05:46.035 * Redis version=7.4.1, bits=64, commit=00000000, modified=0, pid=7, just started
7:C 11 Dec 2024 16:05:46.035 * Configuration loaded
7:M 11 Dec 2024 16:05:46.035 * monotonic clock: POSIX clock_gettime
7:M 11 Dec 2024 16:05:46.036 * Running mode=standalone, port=6379.
7:M 11 Dec 2024 16:05:46.038 * Server initialized
7:M 11 Dec 2024 16:05:46.038 * Ready to accept connections tcp
7:M 11 Dec 2024 16:05:46.038 * Ready to accept connections tls
$
```

- Sentinel - is running in TLS mode and monitoring Redis in TLS mode
```bash
 oc logs redis-sentinel-0
Defaulted container "redis-sentinel" out of: redis-sentinel, copy-sentinel-config (init)
7:X 12 Dec 2024 08:00:57.424 * oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
7:X 12 Dec 2024 08:00:57.424 * Redis version=7.4.1, bits=64, commit=00000000, modified=0, pid=7, just started
7:X 12 Dec 2024 08:00:57.424 * Configuration loaded
7:X 12 Dec 2024 08:00:57.425 * monotonic clock: POSIX clock_gettime
7:X 12 Dec 2024 08:00:57.425 * Running mode=sentinel, port=26380.
7:X 12 Dec 2024 08:00:57.431 * Sentinel new configuration saved on disk
7:X 12 Dec 2024 08:00:57.431 * Sentinel ID is 06a494cc28ba97867ffc5ad8a3cc25864ec2e4f4
7:X 12 Dec 2024 08:00:57.431 # +monitor master mymaster 172.30.115.228 6380 quorum 2
7:X 12 Dec 2024 08:00:57.581 * +sentinel sentinel f5814c076e634ab048069888bd0eef3cff15202d 10.131.0.82 26380 @ mymaster 172.30.115.228 6380
7:X 12 Dec 2024 08:00:57.584 * Sentinel new configuration saved on disk
```

### Summary

By following these steps, we configured Redis and Sentinel to communicate over **TLS** in OpenShift cluster. 
This ensures secure communication between the Redis server and Sentinel.
Please see next section for notes about Clients interaction with the Sentinel and Redis servers

Key steps include:
- Setting up certificates for Redis and Sentinel.
- Configuring Redis and Sentinel to use TLS with the proper configuration files (`redis.conf` and `sentinel.conf`).
- Using OpenShift secrets to securely store certificates and keys.
- Ensuring your Redis clients (either inside or outside OpenShift) are configured to communicate using TLS.  
  


## The Process of Client interaction with the Sentinel and Redis server 

In a **Sentinel** environment, the client application does **not** directly connect to the **Sentinel** nodes. 
Instead, the client application connects to the **Redis server** (e.g., the `redis` server or the current master Redis instance) 
for actual data operations (reads and writes). However, the client needs to interact with the Sentinel
to determine which Redis server is currently the **master** in case of failover or discovery.

Here are more details on how the client application interacts with the Sentinel and Redis server:

### Role of Sentinel
Sentinel is primarily responsible for:
- **Monitoring** the Redis servers (e.g., `redis`).
- **Failover management**: In case the master Redis server goes down, Sentinel promotes one of the replicas to become the new master.
- **Notification**: It informs clients about the current master server.

### How Clients Interact with Sentinel and Redis Servers
- **Initial Connection**: The client application does **not** connect directly to the Sentinel nodes for data operations. Instead, it first connects to the **Sentinel** to determine which Redis server is currently the **master**.
- **Getting the Current Master**: The client sends a `SENTINEL get-master-addr-by-name mymaster` request to the Sentinel to retrieve the address (IP and port) of the **current master** Redis instance. The Sentinel provides this information based on the monitoring data it has.
  - The client sends this query to **any Sentinel** node, and the Sentinel responds with the IP address and port of the current master.
- **Connecting to the Redis Master**: After obtaining the master information from the Sentinel, the client connects directly to the **Redis server** (the current master) for actual data operations (e.g., `SET`, `GET`, etc.).
- **Failover Handling**: If the client is connected to a master Redis instance and that master goes down, Sentinel will promote one of the replica instances to be the new master. The client can then query Sentinel again to get the updated address of the new master. This allows the client to continue its operations without manual intervention.

### How Failover Works with Sentinel
In case of a failover (for example, when the current master goes down), Sentinel promotes a replica to be the new master. 
The client is expected to:
- **Detect Master Change**: The client may receive a notification or keep querying Sentinel to get the address of the current master.
- **Re-establish Connection**: Upon receiving the new master’s address from Sentinel, the client automatically switches 
to the new master and resumes data operations.

Sentinel does not handle actual data requests. It only provides information about the master server and 
handles the failover process. The client needs to be able to handle the transition to a new master automatically.


### Summary
- **Client connects to Sentinel** to discover the current master.
- **Client connects to the current master Redis server** for data operations (reads/writes).
- **In case of failover**, the client queries Sentinel again to get the new master’s address and re-establish the connection.

So, **client applications never directly communicate with the Sentinel** for data operations. 
They always connect to the Redis server (master) that Sentinel has informed them about. 
The role of Sentinel is to monitor Redis instances, handle failover, and provide the current master’s address to the client.  
Please see [operator-user-guide.md](./operator-user-guide.md) for SENTINEL configuration in 3scale.  

- We found it challenging to observe Sentinel's failover process in Kubernetes, as K8S/OpenShift resolves the issue faster than Sentinel. Therefore, we tested Sentinel's behavior on EC2, as detailed in the next section.

## Sentinel Test on EC2 (briefly)
The goal of this test on EC2 was to observe the failover process with Sentinel. As mentioned earlier, it's challenging to 
observe Sentinel's failover process in Kubernetes because K8S/OpenShift handles the issue faster than Sentinel 
can decide which replica to promote as the Redis master.  
We use non-TLS configuration for simplicity.  

Following was done:
- EC2 Ubuntu instance was created
- Redis sources were downloaded, compiled, installed, run.
- Sentinel's failover handling of redis was observed.  
- We are providing only logs here. For installation details of Redis+Sentinel on EC2 - please see AWS and Redis docs or ChatGPT.

- **Redis and Sentinel processes are running**
```shell
ubuntu@ip-10-0-0-53:~$ ps -ef |grep redis
ubuntu     41069    1853  0 16:57 pts/0    00:00:03 /home/ubuntu/redis-stable/src/redis-server 127.0.0.1:6379
ubuntu     41084    2704  0 17:03 pts/2    00:00:02 /home/ubuntu/redis-stable/src/redis-server 127.0.0.1:6380
ubuntu     41295   41190  0 17:13 pts/1    00:00:01 redis-sentinel *:26379 [sentinel]
```

- **Sentinel logs on terminal**

This log provides a detailed view of how Sentinel handles failover events, including master detection, slave promotion, 
and reconfiguration of the Redis nodes. The log entries highlight the exact steps Sentinel takes to 
ensure high availability and smooth operation during Redis failovers.

```bash
ubuntu@ip-10-0-0-53:~$ redis-sentinel redis-sentinal/sentinel.conf
......               _._                                                  
           _.-``__ ''-._                                             
      _.-``    `.  `_.  ''-._           Redis Community Edition      
  .-`` .-```.  ```\/    _.,_ ''-._     7.4.1 (00000000/1) 64 bit
 (    '      ,       .-`  | `,    )     Running in sentinel mode
 |`-._`-...-` __...-.``-._|'` _.-'|     Port: 26379
 |    `-._   `._    /     _.-'    |     PID: 41295
  `-._    `-._  `-./  _.-'    _.-'                                   
 |`-._`-._    `-.__.-'    _.-'_.-'|                                  
 |    `-._`-._        _.-'_.-'    |           https://redis.io       
  `-._    `-._`-.__.-'_.-'    _.-'                                   
 |`-._`-._    `-.__.-'    _.-'_.-'|                                  
 |    `-._`-._        _.-'_.-'    |                                  
  `-._    `-._`-.__.-'_.-'    _.-'                                   
      `-._    `-.__.-'    _.-'                                       
          `-._        _.-'                                           
              `-.__.-'                                               

41295:X 15 Dec 2024 17:13:03.498 * Sentinel new configuration saved on disk
41295:X 15 Dec 2024 17:13:03.498 * Sentinel ID is 385306d9931a6efdde590265c6679840179f2530
41295:X 15 Dec 2024 17:13:03.498 # +monitor master mymaster 127.0.0.1 6379 quorum 2
41295:X 15 Dec 2024 17:13:03.500 * +slave slave 127.0.0.1:6380 127.0.0.1 6380 @ mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:13:03.504 * Sentinel new configuration saved on disk
41295:X 15 Dec 2024 17:17:56.828 # +sdown master mymaster 127.0.0.1 6379

41295:X 15 Dec 2024 17:41:22.173 * Executing user requested FAILOVER of 'mymaster'
41295:X 15 Dec 2024 17:41:22.173 # +new-epoch 1
41295:X 15 Dec 2024 17:41:22.173 # +try-failover master mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:22.266 * Sentinel new configuration saved on disk
41295:X 15 Dec 2024 17:41:22.266 # +vote-for-leader 385306d9931a6efdde590265c6679840179f2530 1
41295:X 15 Dec 2024 17:41:22.266 # +elected-leader master mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:22.266 # +failover-state-select-slave master mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:22.329 # +selected-slave slave 127.0.0.1:6380 127.0.0.1 6380 @ mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:22.329 * +failover-state-send-slaveof-noone slave 127.0.0.1:6380 127.0.0.1 6380 @ mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:22.384 * +failover-state-wait-promotion slave 127.0.0.1:6380 127.0.0.1 6380 @ mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:23.414 * Sentinel new configuration saved on disk
41295:X 15 Dec 2024 17:41:23.414 # +promoted-slave slave 127.0.0.1:6380 127.0.0.1 6380 @ mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:23.414 # +failover-state-reconf-slaves master mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:23.488 # +failover-end master mymaster 127.0.0.1 6379
41295:X 15 Dec 2024 17:41:23.488 # +switch-master mymaster 127.0.0.1 6379 127.0.0.1 6380
41295:X 15 Dec 2024 17:41:23.488 * +slave slave 127.0.0.1:6379 127.0.0.1 6379 @ mymaster 127.0.0.1 6380
41295:X 15 Dec 2024 17:41:23.492 * Sentinel new configuration saved on disk

41295:X 15 Dec 2024 17:41:53.497 # +sdown slave 127.0.0.1:6379 127.0.0.1 6379 @ mymaster 127.0.0.1 6380
```
**Log Explanation**:

The terminal log below shows how Redis Sentinel handled the Redis failover process. Here’s a breakdown of key events observed in the log:

1. **Initial Configuration**:
    - Sentinel starts and loads its configuration, identifying itself with a unique Sentinel ID.
    - Sentinel begins monitoring the Redis master (`mymaster`) at `127.0.0.1:6379`.
    - The slave Redis at `127.0.0.1:6380` is added to the monitoring configuration.

2. **Master Failure Detection**:
    - Sentinel detects that the Redis master (`127.0.0.1:6379`) is down (`+sdown` event) at **17:17:56.828**.

3. **Manual Failover Request**:
    - A user requests a failover for the `mymaster` Redis instance.
    - Sentinel starts the failover process at **17:41:22.173** by selecting a new leader for the failover decision.
    - Sentinel votes for itself as the leader of the failover process (`+vote-for-leader` and `+elected-leader`).

   Note: Please note that at least two Sentinel instances are required for automatic failover in the default configuration. 
   In our example, we had only one Sentinel instance, so to initiate the failover manually, the following command was used:
    ```bash
     redis-cli -p 26379 SENTINEL failover mymaster
    ```
4. **Failover Process**:
    - Sentinel selects the slave (`127.0.0.1:6380`) to be promoted to master (`+selected-slave`).
    - Sentinel issues the command to the selected slave to promote it to the new master (`+failover-state-send-slaveof-noone`).
    - Sentinel waits for the promotion to complete (`+failover-state-wait-promotion`).
    - After successful promotion, the new master (`127.0.0.1:6380`) is confirmed (`+promoted-slave`).

5. **Reconfiguration and Final Switch**:
    - Sentinel reconfigures the remaining slaves to recognize the new master (`+failover-state-reconf-slaves`).
    - The failover process is completed with the new master (`127.0.0.1:6380`) taking over from the failed master (`+switch-master`).

6. **Slave Failure**:
    - Later, Sentinel detects that one of the slaves (127.0.0.1:6379), which was previously the master, is down (+sdown). 
This indicates that the failover and reconfiguration processes are functioning as expected, as the former master is now a slave after the failover.



- **Sentinel Log before failover simulation**
  - **master**

```bash
~ ssh -i ~/.ssh/sentinel1.pem ubuntu@3.145.123.109
Welcome to Ubuntu 24.04.1 LTS (GNU/Linux 6.8.0-1018-aws x86_64)
.....
Last login: Sun Dec 15 17:04:07 2024 from 77.126.86.222
ubuntu@ip-10-0-0-53:~$ redis-cli -p 26379
127.0.0.1:26379> sentinel masters
1)  1) "name"
2) "mymaster"
3) "ip"
4) "127.0.0.1"
5) "port"
6) "6379"
7) "runid"
8) "e4a113cd309302dbb36b20859db876ebfa7ae9f6"
9) "flags"
10) "master"
11) "link-pending-commands"
12) "0"
13) "link-refcount"
14) "1"
15) "last-ping-sent"
16) "0"
17) "last-ok-ping-reply"
18) "268"
19) "last-ping-reply"
20) "268"
21) "down-after-milliseconds"
22) "30000"
23) "info-refresh"
24) "9223"
25) "role-reported"
26) "master"
27) "role-reported-time"
28) "89423"
29) "config-epoch"
30) "0"
31) "num-slaves"
32) "1"
33) "num-other-sentinels"
34) "0"
35) "quorum"
36) "2"
37) "failover-timeout"
38) "180000"
39) "parallel-syncs"
40) "1"
    127.0.0.1:26379> exit  
```
- **slave**

```bash
ubuntu@ip-10-0-0-53:~$ redis-cli -p 26379
127.0.0.1:26379> sentinel replicas mymaster
1)  1) "name"
    2) "127.0.0.1:6380"
    3) "ip"
    4) "127.0.0.1"
    5) "port"
    6) "6380"
    7) "runid"
    8) "f2773429509a3af08b5e2e14b5123cd281839048"
    9) "flags"
   10) "slave"
.....
   35) "master-port"
   36) "6379"
.....
127.0.0.1:26379> 
```

- **After deletion of master redis (fails simulation)**

```bash
ubuntu@ip-10-0-0-53:~$ redis-cli -p 26379
127.0.0.1:26379> 
127.0.0.1:26379> sentinel masters
1)  1) "name"
    2) "mymaster"
    3) "ip"
    4) "127.0.0.1"
    5) "port"
    6) "6380"
    7) "runid"
    8) "f2773429509a3af08b5e2e14b5123cd281839048"
    9) "flags"
   10) "master"
 .......
127.0.0.1:26379> 
```
**Note**: Pay attention to port **6380**, which is the port of the slave Redis. The failover process switched the master from the failed instance to the slave, which is now the new master.

```shell
$ redis-cli -p 26379
127.0.0.1:26379> sentinel replicas mymaster
1)  1) "name"
    2) "127.0.0.1:6379"
    3) "ip"
    4) "127.0.0.1"
    5) "port"
    6) "6379"
    7) "runid"
    8) ""
    9) "flags"
   10) "s_down,slave,disconnected"
   ,,,,,,,,,
```

## Some useful commands

```bash
# Connect to Redis server with TLS
redis-cli -h redis -p 6379 --tls --cert /path/to/client.crt --key /path/to/client.key --cacert /path/to/ca.crt

# Connect to Sentinel with TLS
redis-cli -h redis-sentinel -p 26379 --tls --cert /path/to/client.crt --key /path/to/client.key --cacert /path/to/ca.crt SENTINEL get-master-addr-by-name mymaster
```
If you're inside an OpenShift pod and want to use `redis-cli` from within a container:
```bash
oc rsh <sentinel-pod-name>
redis-cli -h redis-sentinel -p 26379 --tls --cert /etc/redis/tls/client.crt --key /etc/redis/tls/client.key --cacert /etc/redis/tls/ca.crt SENTINEL get-master-addr-by-name mymaster
```