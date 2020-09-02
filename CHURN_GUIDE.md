# process

* get all accounts
  * cli is `mychurnero get-all-accounts`
* for each account get all subaddresses
  * cli is `mychurnero get-address`
* for each subaddress check if it is used
* ignore unused subaddresses
* for all used subaddresses get spendable balance
* schedule sending in random intervals
  * default to 1hr -> 24hr delay from time of receive
  * allow user to control the time (optional)
* send transaction noting it in sqlite database