## Client
Start the node using 
```bash
go run main.go --port=9000
```
Then simulate simulate multiple connections using client.go

```bash
Example command: go run client.go --network tcp --address ":9000" --concurrency 5 --packet_batch 1 --packet_count 10
```

The client simulator will connect to the first node, and send Announcement packets

## Networking
### Encryption and Hashing functions

* Salsa20 is used for encrypting the packets.
* secp256k1 is used to generate the peer IDs based on the public keys.
* blake3 is used for hashing the packets when signing.

### Network Packet

| Offset        | Content                                                                                |
|---------------|----------------------------------------------------------------------------------------|
| 0:6           | Packet Header: Magic Number + Nonce                                                    |
| 6:14+N        | Packet Body Encrypted: Version + Command + Sequence + Payload Size + Payload + Garbage |
| 14+N: 14+N+65 | Packet Footer Encrypted: Signature                                                     |

### Protocol

When content of bytes [6:14+N+65] is decrypted we get

| Offset | Length | Content                                            |
|--------|--------|----------------------------------------------------|
| 0      | 2      | Magic Number                                       |
| 2      | 4      | Nonce                                              |
| 6      | 1      | Protocol version = 0                               |
| 7      | 1      | Command                                            |
| 8      | 4      | Sequence                                           |
| 12     | 2      | Size of Payload data                               |
| 14     | ?      | Payload                                            |
| ?      | ?      | Randomized garbage                                 |
| ?      | 65     | Signature, ECDSA secp256k1 512-bit + 1 header byte |

#### Announcement

