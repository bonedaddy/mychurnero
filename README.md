# mychurnero

Mychurnero is a monero churning service that allows you to automatically, and randomly churn your monero. It coordinates churning by storing information in a local sqlite3 database, as opposed to using wallet tags which are persisted within the monero wallet. This allows you to securely churn from a remote location, without disclosing possibly identifying tags to the monero-wallet-rpc node. This is experimental software, and the benefits of churning are not conclusive. As such please make sure you read the warnings section down below.

Interested in supporting mychurnero development? Send XMR to `87Gkfh2VPLjJQ1EYkSXzs6GBbY3Arwm1NJKpkyhQzbf7R8iu8VSnjFDC23vaUc5TFK7boZtPV2kXhSQYZenwtWTzPGdRBds`

# terminology

* `churn from`
  * this means the subaddresses we are sending funds from
* `churn to`
  * this means the subaddress we are depositing churned funds into
* `churn account index`
  * this is the account index we use for generating all addresses to deposit churned funds into

# warnings

* loss of funds may occur from using this as it is experimental software
* mychurnero may not provide any benefits at all
* guard access to the sqlite3 database on disk with care, as this can be used to identify churned transactions
  * once done with churning, securely delete the sqlite3 database
  * information is only persisted in the sqlite3 database as long as is needed and the moment a churn transaction is confirmed this information is removed from the database, but do not solely rely on this
* do not use a single mychurnero instance for multiple different wallets
  * there are thread-safety concerns when handling multiple different wallets at the same time
  * the database connection leverages a shared cache which if using multiple different wallets could potentially be a source of information leakage although this isn't serious
* transaction fees are randomly determined and could be costly
  * transaction fee analysis could be used for fingerprinting
* this may or may not relay transactions through anonymized networks such as Tor or I2P however that will entirely depend on the monerod node your monero-wallet-rpc client talks to

# usage

For details usage instructions see [USAGE.md](./USAGE.md)

# why does this need a database

In order to process, schedule, and delay churning transactions there needs to be some way of identifying addresses, as well as persisting information about transactions to be churned in case of power loss, and similar situations. One option to identify addresses for churning was to leverage monero wallet tags, however this means if anyone gets access to the wallet, or you are using a remote monero-wallet-rpc node, they would be able to use this information against you. It would also mean that you always need to use the same monero-wallet-rpc node each time you want to churn. 

To alleviate these privacy concerns, and to allow you to freely use any remote monero-wallet-rpc node, we needed to use a database. Sqlite3 was picked as this allows us easily churn from any location, using any monero-wallet-rpc without disclosing tags to a monero-wallet-rpc node. It also makes it extremely easy to delete the churning information once done, as a simple `shred` on the database file and you are done!

# churnable address selection

The process of selecting an address we can churn from is pretty easy, and simply consists of ensuring that the subaddress does not belong to the account index we use for generating churn to addresses, and that the minimum unlocked balance is equal or greater than the minimum. If a subaddress satisfies this requirement it is then eligible for being churned from.

# where are the funds deposited

When configuring mychurnero you specify an account index called "churn account index", and each time a churn transaction is created, a subaddress under this account will be generated for each churn transaction and never be reused. It is up to you to determine how to sweep and aggregate funds under this account index.


# churning process

Every user configured periodic interval we will scan all account indexes *except* the churn account index for subaddresses with an unlocked balance greater than our specified minimum. If an address passes this check, and it has no other currently scheduled transactions we then generate a transaction but do not relay it. A random delay is picked, and this information along with the transaction metadata is stored in a sqlite3 database. At the end of transaction creation we then check to see if any of the previously created transactions have had their delays past. Any transactions which have past this delay are then relayed and the sqlite entry for this transactions is marked as having been relayed. 

After all eligible transactions have been relayed, we then check previously relayed transactions to determine if they are confirmed. To determine whether or not a transaction is confirmed, we use the `SuggestedConfirmationsThreshold` information returned by the monero-wallet-rpc node being used. If a transaction has at least this many confirmation it is considered confirmed, and all information relayed to this transaction and its associated churn from address are removed from the sqlite3 database.


# links

* https://www.reddit.com/r/Monero/comments/egxulr/we_need_better_ways_to_combine_multiple_outputs/fcbakt6/
* https://www.reddit.com/r/Monero/comments/ekz7wg/what_is_the_latest_consensus_on_minimum_amount_of/
* https://www.reddit.com/r/Monero/comments/a6b3ea/whats_the_latest_consensus_on_minimum_amount_of/
* https://www.reddit.com/r/Monero/comments/70zalp/clarification_on_mrls_churning_comments/dn7a7wa/
* https://monero.stackexchange.com/questions/7460/cant-transfer-monero-coins-strange-error-message-error-not-enough-outputs-fo