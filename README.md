<a href="https://peervault.github.io/architecture-rfcs">
  <img src="https://peervault.github.io/architecture-rfcs/images/peervault_logo.svg" alt="PeerVault" width="80px" align="left" style="padding:25px 20px 0 20px"/>
</a>

> **PeerVault** is a peer-to-peer decentralized application used to share sensitive information with someone else identified by cryptographic keys. 
The Vault is yours, no server involved, **100% Open Source**.
Leverage from blockchain technology innovation such as libp2p and bip39 Mnemonic code for generating deterministic keys.

# PeerVault Service (Daemon)

:warning: if you are interested about using the program, you should head to the [CLIENT APP](https://github.com/PeerVault/PeerVault-GUI-Flutter) that already includes the service in the same bundle.

## QuickStart

Download the binary either for Linux or OSX depending of your platform on the github release page

- [Download for OSX](https://github.com/PeerVault/PeerVault-Service/releases/download/v0.1.0/peervault-darwin)
- [Download for Linux](https://github.com/PeerVault/PeerVault-Service/releases/download/v0.1.0/peervault-linux)

### Usage

```
❯ ./bin/peervault --help                                                                                                                                                     PeerVault-Service/git/master !
Usage of ./bin/peervault:
  -apiAddr string
    	http api service address (default "localhost:4444")
  -bbolt string
    	Location of bbolt DB file
  -dev
    	Enable dev mode
  -log int
    	Log level, 3=error, 6=notice, 9=debug
  -logfile string
    	Location of log file
  -relay string
    	Relay Host URL
  -wsAddr string
    	WebSocket event service address (default "localhost:5555")
```

Start the daemon in foreground with debug log enabled

```
❯ ./bin/peervault \
    --log 9 \
    --relay /ip4/37.187.1.229/tcp/23003/ipfs/QmeFecyqtgzYx1TFN9vYTroMGNo3DELtDZ63FpjqUd6xfW \
    --bbolt ~/bbolt.db
!!! ATTENTION !!!
DEBUG LOGGING MAY CONTAIN SENSIBLE INFORMATION SUCH AS CLEAR PRIVATE KEY OR ANY DATA.
IT SHOULD ONLY BE USED IN DEVELOPPER MODE

19:01:05.523 Listen ▶ DEBU 002 Listen
19:01:05.523 Listen ▶ INFO 001 listen from event
19:01:05.524 Listen ▶ INFO 003 listen from control
19:01:05.525 CreateOrOpen ▶ DEBU 004 Checking keychain status
19:01:05.551 CreateOrOpen ▶ DEBU 005 Keychain status returned nil, keychain exists
19:01:05.552 Get ▶ DEBU 006 16Uiu2HAm5fREw7TtUEjkru7xgHZGpRTLRxrtVVMFmxKNZgX4qRFr Owner
19:01:05.822 Listen ▶ INFO 007 listen from peer
19:01:05.822 Listen ▶ INFO 008 16Uiu2HAm5fREw7TtUEjkru7xgHZGpRTLRxrtVVMFmxKNZgX4qRFr
19:01:05.822 Listen ▶ INFO 009 [/ip4/127.0.0.1/tcp/50451 /ip4/127.94.0.1/tcp/50451 /ip4/192.168.127.155/tcp/50451 /ip6/::1/tcp/50452]
```

### Functional

If you wish to test all the functionaly without the GUI
You can use POSTMAN to request the service.
Import the collection using `api-postman-collection.json`

## Roadmap

### MVP - Phase 2 (Q2 2020)

### MVP - Phase 1 (Q1 2020)

- [x] API CRUD `/owner` (Create new Vault Owner)
- [x] API POST `/owner/seed` (Recovery from Paper Key)
- [x] API CRUD `/secret` (Manage secrets)
- [x] API CRUD `/expose/request` (Share secrets with others Peer)
- [x] WebSocket notification when sharing request received
- [x] Sharing secrets protocol ([View Doc Architecture](https://peervault.github.io/architecture-rfcs/architecture/protocol/secret-sharing.html))
- [ ] Minimal GUI to retrieve information ([View Flutter GUI](https://github.com/PeerVault/PeerVault-GUI-Flutter))

### PoC - Phase 0 (Q1 2019)

- [x] Cryptography Key Derivation BIP32 / BIP39
- [x] Peer Libp2p with Relay integration ([View Relay example](https://github.com/PeerVault/go-libp2p-relay-app))
- [x] Exchange string between Peer through NAT using relay

### Quality integration

- [x] Security by design, no server involve to make it work
- [x] Strong private key generation using bip39 / bip32
- [x] Secure private key inside OSX Keychain
- [x] Encryption of all secrets using private key
- [x] Logging system in place for easily debuging and print error
- [x] Clean directory tree structure for clean software architecture
- [x] Api / Peer / Notification are split in three server, HTTP, LIBP2P, WS
