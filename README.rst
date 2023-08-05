=====
Penny
=====

Penny is personal finance tracking web application.

-------------
Install & Run
-------------

The quickest and easiest way to run Penny is to build and run the docker container:

    $ task docker:build docker:run

Then visit http://localhost:5000 in your browser.

If you want to build and run Penny from source, first start redis on localhost:6379.

Then build::

    $ task build

And run the web app::

    $ task run:www

And finally, in another terminal, run the queue system::

    $ task run:queue

The queue system is backed by redis and defaults to connecting to redis on localhost:6379.

------------
Dependencies
------------

A list of hard dependencies that are needed when building Penny:
 * python 3.9:
   - Current build system expects python 3.9. Some of the python packages used
     by penny are failing to build with the latest version of python (3.11 at
     time of writing). If debian bookworm is being used, the use of python3.9
     is enforced when running `task build`.

-----------
Get Started
-----------

Visit http://localhost:5000/register and register your first account.

Once you've logged in, create an Entity. Entities are Companies, Sole Traders or a Person.

Next, create a Bank Account. A Bank Account has a Type such as Savings or Credit Card. A Bank Account belongs to an Entity.

Next, create a Transaction. A Transaction is linked to a Bank Account, which in turn is linked to an Entity.

A Transaction can also be linked to an Account. Accounts have a Type such as Expense or Liability. Accounts are also linked to Entities.

Transactions can be automatically matched to Accounts by Account Matches. Account Matches apply filters to link Transactions to Accounts.

The only type of report that is currently supported is a Profit and Loss
report. P&L reports are automatically created for Entities and Bank Accounts.
