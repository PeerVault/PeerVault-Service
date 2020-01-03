## PeerVault - P2P Secret Vault

## PoC Roadmap - Phase 0

- [ ] Http API Create owner / secret / share
- [x] Peer Libp2p with Relay integration
- [ ] WS Notification new share received
- [ ] minimal GUI to retrieve information
- [ ] Share secrets with other Peer, Libp2p / Ws (involve)

## Quality integration

- [x] Security by design, no server involve to make it work
- [x] Strong private key generation using bip39 / bip32
- [x] Secure private key inside OSX Keychain
- [x] Encryption of all secrets using private key
- [x] Logging system in place for easily debuging and print error
- [x] Clean directory tree structure for clean software architecture
- [x] Api / Peer / Notification are split in three server, HTTP, LIBP2P, WS