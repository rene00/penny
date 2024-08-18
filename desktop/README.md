# penny desktop

See [README.md](../README.md) in parent directory.

## TODO

`BUG` or `FEAT` todo items.

* [ ] BUG: `transaction_importer_status.status_id` is set to `0` once
      successful auto transaction has completed.
* [ ] BUG: when adding new filter within UI, "add new filter" button is not
      disabed after filter is available.
* [ ] BUG: opening attachment only works with operating systems that have
      xdg-open installed in /usr/bin/open. This needs to be extended to
      support other operating systems.
* [ ] BUG: importing OFX transactions will fail if currency symbol within ofx
      are invalid.
* [ ] FEAT: transaction attachments should have a size limit
* [ ] FEAT: support deleting transaction attachments
* [ ] FEAT: account matches should not have to be duplicated when matching on
      destination account across multiple source accounts. For example,
      matching for groceries across 2 credit accounts should be a single
      account match, not 2 account matches.
* [ ] BUG: tag amount is incorrectly calculated. Build a tag with many
      transactions and confirm that total sum amount is correct.
