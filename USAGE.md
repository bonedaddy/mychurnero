# Usage

This guide covers usage, and installation of mychurnero. Obviously before using you will need to install, so please see the installation menu below.

# Installation

At the moment the only supported method of installing mychurnero is building from source. You will need Golang installed with a minimum supported version being 1.14.x, as well as git installed. Once those two dependencies are installed, please run the following commands

```shell
$> git clone https://github.com/bonedaddy/mychurnero.git
$> cd mychurnero
$> make # builds file named mychurnero in current directory
```

# Configuration

Mychurnero uses a yaml configuration file to control the churning process. You have a few different options for doing this:

Placing the configuration file in the default location of `mychurnero.yml`:

```shell
$> mychurnero config-gen
```

Placing the configuration file in the location of `/tmp/mychurnero.yml`:

```shell
$> mychurnero -config /tmp/mychurnero.yml config-gen
```

Note that if you have your config file in a location other than the default, you will need to specify its path using `-config` flag whenever invoking the service command. The default configuration file is depicted below, along with comments explaining the options

```yaml
# this is the path to store the sqlite database
dbpath: mychurnero.db
# this is the name of the wallet we want to use
# please note that it must be accessible by the monero-wallet-rpc node being used
walletname: testnetwallet123
# this is the address of the monero-wallet-rpc endpoint
rpcaddress: http://127.0.0.1:6061/json_rpc
# the name of the file to store logs in
# this may contain sensitive information
logpath: mychurnero.log
# this is the account index we use to generate subaddresses to deposit churned funds into
# any subaddresses under this account index will never be churned from
churnaccountindex: 1
# this defines the minimum amount of murnero to churn, this number means `0.1` monero
# to convert a decimal number to the corresponding uint64 monero value run the `mychurnero covnert-to-xmr` command
minchurnamount: 100000000000
# this is the minimum delay in minutes to use for scheduling transactions
mindelayminutes: 1
# this is the maximum delay in minutes to use for scheduling transactions
maxdelayminutes: 10
# specifies the frequency for which we will look for new addresses we can churn from
scaninterval: 2m25s
```