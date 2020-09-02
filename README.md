# mychurnero

mychurnero is a monero churning service that allows you to automatically, and randomly churn your monero! In theory monero churning allows you to mix your funds and get better plausible deniability, always as reduced linkability to previous transactions.

Interested in supporting mychurnero? Send XMR to `87Gkfh2VPLjJQ1EYkSXzs6GBbY3Arwm1NJKpkyhQzbf7R8iu8VSnjFDC23vaUc5TFK7boZtPV2kXhSQYZenwtWTzPGdRBds`

# what defines a churnable address

a churnable address is defined as one that has previously been used to receive a transaction, and that has an unlocked balance greater than 0. right now we use a pretty naive method of retrieving this information which is essentially parsing over all available wallet and account information. This means if you have a wallet with multiple accounts and each account has multiple sub addresses it will take a lot longer to retrieve the churnable information than if you had one account with one subaddress.

In the future we will use a sqlite database to cache this information so each time we need to retrieve churnable address information, we can start off with the previous state, and parse that instead of starting from scratch each and every time.

# how often will churning happen

to prevent churning to frequently, the default setting is to churn within 1hr -> 24hrs after an address last received a transaction. the end goal is that no two churn transactions will be broadcast at the same time, and the tx fee, as well as amounts ent will be varied. for now we will take a pretty naive approach of queueing all churn transactions within the predefined bounds, using the same transaction fee, and the amount that is unlocked for a given address whenever a churn is started.