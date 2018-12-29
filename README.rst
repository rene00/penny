=====
Penny
=====

Penny is personal and micro business finance tracking web application.

-------------
Install & Run
-------------

Start redis.

Build::

    $ make build

Run web app::

    $ make run

In another terminal, run the queue system::

    $ make run_queue

Visit http://localhost:5000 in your browser.

-----------
Get Started
-----------

Visit http://localhost:5000/register and register your first account.

Entities are companies, sole traders or people. Bank Accounts belong to
entities. Transactions are linked to Bank Accounts. Transactions can also be
linked to Accounts, which are categories of Transactions.  Account Matches
apply filters to link Transactions to Accounts.

The only type of report that is currently supported is a Profit and Loss
report. P&L reports are automatically created for Entities and Bank Accounts.
