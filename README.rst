=====
Penny
=====

Penny is finance tracking web application.

-------------
Basic Install
-------------

Install and run redis::

    $ sudo apt-get install redis-server

Install and run mariadb::

    $ sudo apt-get install mariadb-server

Setup a MySQL account and table for penny::

    $ mysql
    ...
    > CREATE DATABASE penny;
    > CREATE USER 'penny'@'localhost'
        IDENTIFIED BY 'penny';
    > GRANT ALL PRIVILEGES ON penny.*
        TO 'penny'@'localhost';
    > FLUSH PRIVILEGES;
    > QUIT;

Clone the latest version of penny::

    $ git clone https://github.com/rene00/penny
    $ cd penny

Build penny::

    $ make

Run the database migration scripts::

    $ make db_migrate

Run penny::

    $ make run_www

In another terminal, run the queue system::

    $ make run_queue

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
