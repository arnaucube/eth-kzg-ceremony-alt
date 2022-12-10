# eth-kzg-ceremony-alt [![GoDoc](https://godoc.org/github.com/arnaucube/eth-kzg-ceremony-alt?status.svg)](https://godoc.org/github.com/arnaucube/eth-kzg-ceremony-alt) [![Test](https://github.com/arnaucube/eth-kzg-ceremony-alt/workflows/Test/badge.svg)](https://github.com/arnaucube/eth-kzg-ceremony-alt/actions?query=workflow%3ATest)

Alternative (non-official) implementation in Go of the *contributor* for the [Ethereum KZG Trusted Setup Ceremony](https://github.com/ethereum/kzg-ceremony/blob/main/FAQ.md).

The purpose of this repo is to use it to contribute to the upcoming Ethereum KZG Trusted Setup Ceremony, without using the official implementation.

The Ceremony is considered safe as long as there is at least one honest participant, with the idea that if you participate, assuming that you consider yourself honest, you can consider the Ceremony safe.
Ethereum will run the Ceremony which will be used at least in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844).

Probably most of the contributions will be generated with the same code (official impl, which has been audited). The idea of this repo is to try to bring more diversity to the table with another independent implementation.

This implementation has been done without looking at the other impls code (except for the point parsers test vectors), in order to not be biased by that code.

> This code has not been audited, use it at your own risk.

Why in Go? Ideally would have done this code using Rust & arkworks, but the official impl already uses that. This implementation uses [Kilic's BLS12-381 pairing implementation](https://github.com/kilic/bls12-381).

Documents used for this implementation:
- [KZG10-Ceremony-audit-report.pdf, section *3.1 Overview of PoT ceremonies*](https://github.com/ethereum/kzg-ceremony/blob/main/KZG10-Ceremony-audit-report.pdf)
- [*Why and how zkSNARKs work*, by Maksym Petkus](https://arxiv.org/abs/1906.07221v1)

### Usage
Get the binary from the [releases](https://github.com/arnaucube/eth-kzg-ceremony-alt/releases) (alternative you can compile it from source), and run:
```
> ./kzgceremony

eth-kzg-ceremony-alt
====================

Usage of ./kzgceremony:
  -u, --url string       sequencer url (default "https://kzg-ceremony-sequencer-dev.fly.dev")
  -r, --rand string      randomness
  -s, --sleeptime uint   time (seconds) sleeping before trying again to be the next contributor (default 10)
```

So for example, run your contribution with:
```
./kzgceremony -r "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod"
```
(where the "Lorem ipsum..." is your source of randomness)
