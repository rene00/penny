# penny

penny is an **experimental** personal finance tracking web and desktop
application.

## Current Status

**WARNING**: penny is **experimental**.

It's missing alot of features that you would expect in a personal finance
tracking application. I continue to work on it during my free time but would
recommend using other free and open applications such as
[gnucash](https://gnucash.org/).

There is a [web](./web) and [desktop](./desktop) version. I'm focused on
developing the desktop version with the web version mostly only receiving
maintenance fixes.

## User Documentation

Documentation for the desktop application. Check out [web](./web) for
documentation for web application.


### Install

Download the penny binary the
[latest](https://github.com/rene00/penny/releases/latest) releases page.

Currently only linux amd64 is supported.

### Importing Transactions

Only OFX files for Credit Card Statements and Checking Accounts are currently
supported.

When transactions are imported, the credit card or checking account will be
automatically created with the new transaction being automatically assigned to
this account.

#### Auto Importing Transactions

OFX files found within `DATA_HOME/import` will be automatically imported when
`penny` is running. On Linux, `DATA_HOME` is generally
`$HOME/.local/share/pennyapp`.

Files that have been successfully imported will be deleted from
`DATA_HOME/import`.
