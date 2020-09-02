# starting monero private network

## start fresh network

run `./setup_testnet.sh` to create the private network if you want to start from scratch

## run the node

```shell
$> monerod --testnet --no-igd --hide-my-port --data-dir testnet/node1/datadir --fixed-difficulty 100 --offline
```
## start mining

after node has started run
```
start_mining 9wviCeWe2D8XS82k2ovp5EUYLzBt9pYNW2LXUFsZiv8S3Mt21FZ5qQaAroko1enzw3eGr9qC7X1D7Geoo2RrAotYPwq9Gm8 1
 ```