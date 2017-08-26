=====
Penny
=====

Penny is personal and micro business finance tracking web application.

-------------
Basic Install
-------------

Install and run redis::

    # ubuntu 16.04
    $ sudo apt-get install redis-server

    # fedora 26
    $ sudo dnf install redis
    $ sudo systemctl start redis.service
    $ sudo systemctl enable redis.service

Install fabric::

    # fedora 26
    $ sudo dnf install fabric

Install the build dependencies::

    # fedora 26
    $ sudo dnf install openssl-devel libffi-devel sqlite-devel sqlcipher-devel

Clone the latest version of penny::

    $ git clone https://github.com/rene00/penny
    $ cd penny

Setup a virtualenv then build and run penny::

    $ mkvirtualenv penny
    $ fab run

In another terminal, run the queue system::

    $ fab run_queue

Visit http://localhost:5000 in your browser.

------------------------
Setup and Brief Overview
------------------------

Visit http://localhost:5000/register and register your first account.

Entities are companies, sole traders or people. Bank Accounts belong to
entities. Transactions are linked to Bank Accounts. Transactions can also be
linked to Accounts, which are categories of Transactions.  Account Matches
apply filters to link Transactions to Accounts.

The only type of report that is currently supported is a Profit and Loss
report. P&L reports are automatically created for Entities and Bank Accounts.
