# penny desktop

See [README.md](../README.md) in parent directory.

## TODO

### `BUG` 
* [ ] BUG: `transaction_importer_status.status_id` is set to `0` once
      successful auto transaction has completed.
* [ ] BUG: when adding new filter within UI, "add new filter" button is not
      disabed after filter is available.
* [ ] BUG: opening attachment only works with operating systems that have
      xdg-open installed in /usr/bin/open. This needs to be extended to
      support other operating systems.
* [ ] BUG: importing OFX transactions will fail if currency symbol within ofx
      are invalid.
* [ ] BUG: tag amount is incorrectly calculated. Build a tag with many
      transactions and confirm that total sum amount is correct.
* [ ] BUG: when deleting a transaction error dialog is displayed with empty
      message though transaction is successfully deleted. Dialog should display
      successful message instead.
* [ ] BUG: Imbalanced account should be immutable. Its attributes should not be
      allowed to be change via UI.
* [ ] BUG: Imbalanced total amount balance is incorrect. To reproduce, add
      several transactions with no account and sum total amount balance. 

### `FEAT`

* [ ] FEAT: transaction attachments should have a size limit
* [ ] FEAT: Add support for deleting transaction attachments
* [ ] FEAT: account matches should not have to be duplicated when matching on
      destination account across multiple source accounts. For example,
      matching for groceries across 2 credit accounts should be a single
      account match, not 2 account matches.
* [ ] FEAT: tags should require matching on account type. Account type should
      default to expense. If tags do not match transaction splits on account
      type, tagging will match all splits in the transaction which results in a
      $0 amount.
