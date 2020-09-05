# mychurnero

mychurnero is a monero churning service that allows you to automatically, and randomly churn your monero! In theory monero churning allows you to mix your funds and get better plausible deniability, always as reduced linkability to previous transactions.

Interested in supporting mychurnero? Send XMR to `87Gkfh2VPLjJQ1EYkSXzs6GBbY3Arwm1NJKpkyhQzbf7R8iu8VSnjFDC23vaUc5TFK7boZtPV2kXhSQYZenwtWTzPGdRBds`

# warnings

* loss of funds may occur from using this as it is experimental software
* the method of churning hasn't been statistically analyzed to determine effectiveness, it works in theory but practice may be different
* it is not safe to use a single mychurnero instance for multiple different wallets
  * there are thread-safety concerns when handling multiple different wallets at the same time
  * the database connection leverages a shared cache which if using multiple different wallets could potentially be a source of information leakage although this isn't serious
* transaction fees are randomly determined and could be costly
* this may or may not relay transactions through anonymized networks such as Tor or I2P however that will entirely depend on the monerod node your monero-wallet-rpc client talks to


# what defines a churnable address

a churnable address is defined as one that has previously been used to receive a transaction, and that has an unlocked balance greater than 0. right now we use a pretty naive method of retrieving this information which is essentially parsing over all available wallet and account information. This means if you have a wallet with multiple accounts and each account has multiple sub addresses it will take a lot longer to retrieve the churnable information than if you had one account with one subaddress.

In the future we will use a sqlite database to cache this information so each time we need to retrieve churnable address information, we can start off with the previous state, and parse that instead of starting from scratch each and every time.

# how often will churning happen

to prevent churning to frequently, the default setting is to churn within 1hr -> 24hrs after an address last received a transaction. the end goal is that no two churn transactions will be broadcast at the same time, and the tx fee, as well as amounts ent will be varied. for now we will take a pretty naive approach of queueing all churn transactions within the predefined bounds, using the same transaction fee, and the amount that is unlocked for a given address whenever a churn is started.

# churning process

Before churning you create an account within a wallet specifically for receiving churned funds. All subaddresses created under this account will never be churned from only churned to, that is we will never send transactions containing funds from that account only send funds to it. Make note of the account index.

Every 20 minutes we scan the wallet, ignoring the churn account, and see if any funds are unlocked. Any accounts who have unlocked funds above the minimum requirement will have a transaction scheduled that will spend all available unlocked funds. This transaction will be broadcast within a predefined window, randomly scheduled between the lower and upper bounds.

# links

* https://www.reddit.com/r/Monero/comments/egxulr/we_need_better_ways_to_combine_multiple_outputs/fcbakt6/
* https://www.reddit.com/r/Monero/comments/ekz7wg/what_is_the_latest_consensus_on_minimum_amount_of/
* https://www.reddit.com/r/Monero/comments/a6b3ea/whats_the_latest_consensus_on_minimum_amount_of/
* https://www.reddit.com/r/Monero/comments/70zalp/clarification_on_mrls_churning_comments/dn7a7wa/