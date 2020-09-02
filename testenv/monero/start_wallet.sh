#! /bin/bash

monero-wallet-rpc --testnet --disable-rpc-login --prompt-for-password --rpc-bind-port 6061 --wallet-dir=testnet/node1

# monero-wallet-rpc --testnet --trusted-daemon --wallet-file testnet/node1/wallet_01.bin --log-file testnet/node1/wallet_01.log