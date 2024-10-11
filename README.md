# Ikto

Ikto is a wireguard mesh builder based on nats. The first Ikto goal is to be the building block of our micro-vm orchestrator, [Ravel](https://github.com/valyentdev/ravel), networking features.

## Concepts

Ikto connects to a [NATS Jetstream](https://docs.nats.io/nats-concepts/jetstream) KV bucket and watch it. It update the local peer configuration and distant peers when changes are made on the KV bucket. In fact, the peer authentication is made via NATS. **If a node has an authenticated read-write access to the NATS cluster and the KV store, it can add himself to the mesh.**
This means that, for now, there is now central control plane for the mesh.

### IPAM 
It's important that each node get an unique IP address. Here are the steps that Ikto follow to reach this goal:
1. Before starting ikto, you generate a random address with `ikto init`.
2. When ikto start, it gets the value on the key `peers.{base64_encoded_peerIP}`.
3. If a value already exist, it checks that the corresponding peer is himself (comparing the public keys) and update itself
4. If not ikto fails because of the already in use address and you need to generate a new random IP (so Ikto will better work if you have a lot more available ip than nodes)
5. If it doesn't exist, Ikto try to create a value [with optimistic locking ](https://docs.nats.io/nats-concepts/jetstream/key-value-store#atomic-operations-used-for-locking-and-concurrency-control)

As a consequence of nats concurrency control properties, duplicated addresses should never happend. 


## Getting started

### Pre-requisites
- An available NATS cluster (or just one nats-server)
- Wireguard installed on each node


### Installation

You can download the latest release from github releases:
```bash
wget https://github.com/valyentdev/ikto/releases/download/v0.3.0/ikto_0.3.0_linux_amd64.tar.gz
tar -xvf ikto_0.3.0_linux_amd64.tar.gz
cp ikto SOMEWHERE_IN_YOUR_PATH
```

We'll provide an install script in the future.

### Configuration

On each node, you can configure ikto:
```bash
$ ikto init > ikto.json
```

File generated:
```json
{
  "name": "",
  "advertise_address": "",
  "private_address": "fd10:2082:5bc1::",
  "subnet_prefix": 48,
  "mesh_cidr": "fd10::/16",
  "wg_dev_name": "wg-ikto",
  "wg_port": 51820,
  "private_key_path": "",
  "nats_creds": "",
  "nats_url": "nats://",
  "nats_kv": "ikto-mesh"
}
```

Finally you can run it:
```bash
$ ikto agent -c ikto.json
```


Ikto listen on an unix socket by default on /tmp/ikto.sock 
```bash
$ ikto agent -c ikto.json -s /var/run/ikto.sock
```


## Contributing

You can signal bugs or request a feature by opening an issue and/or a pull request on this repository. If you have any question you can join our [Discord](https://discord.valyent.dev/) where we are available almost every days. 

## License

   Copyright 2024 SAS Valyent

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use these files except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
