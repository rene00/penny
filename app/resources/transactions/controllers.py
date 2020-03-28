from app import models
from app.common import forms, tasks
from app.common.currency import to_cents, get_credit_debit
from flask import (Blueprint, abort, g, render_template, url_for,
                   current_app as app, send_from_directory, redirect,
                   request, flash)
from flask_security import login_required
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm.exc import NoResultFound
from app.resources.transactions.forms import (FormTransaction,
                                              FormTransactionAdd,
                                              FormTransactionSplit,
                                              FormTransactionUpload)
from app.common.attachments import (get_filepath_for_transaction_attachment,
                                    get_hash_of_file)
import os
from werkzeug import secure_filename
import hashlib
import datetime
import re


transactions = Blueprint('transactions', __name__, url_prefix='/transactions')


def _get_form_split(request, form):
    """Split transaction with FormTransactionSplit().

    If user is creating a new split, return the split form.
    """
    app.logger.debug("Split transaction requested.")
    form_split = FormTransactionSplit()
    form_split.split_account.query = form.account.query
    return form_split


def _save_new_split_transaction(request, transaction):
    """Save new split transaction."""

    split_amount = request.form.get('split_amount')
    split = None

    if split_amount:
        split_account = request.form.get('split_account')
        split_memo = request.form.get('split_memo')
        (split_credit, split_debit) = get_credit_debit(
            to_cents(split_amount))
        app.logger.debug('Saving/Updating Split; account={account} '
                         'memo={memo}, amount={amount}, '
                         'parent_id={transaction.id}'.format(
                             account=split_account, memo=split_memo,
                             amount=split_amount, transaction=transaction))

        # Update the split account if user provided one.
        split = models.Transaction(
            user=g.user, date=transaction.date, memo=split_memo,
            debit=split_debit, credit=split_credit,
            parent_id=transaction.id)

        # Split always inherits the bankaccount of the
        # transaction.
        if transaction.bankaccount:
            split.bankaccount = transaction.bankaccount

        try:
            account = (models.db.session.query(models.Account).
                       filter_by(id=split_account, user=g.user).one())
        except NoResultFound:
            pass
        else:
            app.logger.debug('Found account for split; account_id={0.id}'.
                             format(account))
            split.account = account

        models.db.session.add(split)

    return split


@transactions.route('/')
@login_required
def _transactions():
    return render_template(
        'transactions.html', data_url=url_for('data_transactions.transactions')
    )


@transactions.route('/bankaccount/<int:id>',
                    defaults={'start_date': None, 'end_date': None},
                    methods=['GET', 'POST'])
@transactions.route('/bankaccount/<int:id>/<int:start_date>/<int:end_date>',
                    methods=['GET', 'POST'])
@login_required
def bankaccount(id, start_date, end_date):
    return render_template('transactions.html',
                           data_url=url_for('data_transactions.bankaccount',
                                            id=id, start_date=start_date,
                                            end_date=end_date))


@transactions.route('/account',
                    defaults={
                        'start_date': None, 'end_date': None, 'id': None
                    },
                    methods=['GET'])
@transactions.route('/account/<int:id>',
                    defaults={'start_date': None, 'end_date': None},
                    methods=['GET'])
@transactions.route('/account/<int:id>/<int:start_date>/<int:end_date>',
                    methods=['GET'])
def account(id, start_date, end_date):
    return render_template('transactions.html',
                           data_url=url_for('data_transactions.account',
                                            id=id, start_date=start_date,
                                            end_date=end_date))


@transactions.route(
    '/accounttype/<string:accounttype>',
    defaults={'start_date': None, 'end_date': None},
    methods=['GET']
)
@transactions.route(
    '/accounttype/<string:accounttype>/<int:start_date>/<int:end_date>',
    methods=['GET']
)
def accounttype(accounttype, start_date, end_date):
    return render_template(
        'transactions.html',
        data_url=url_for(
            'data_transactions.accounttype',
            accounttype=accounttype,
            start_date=start_date,
            end_date=end_date
        )
    )


@transactions.route('/<int:id>', methods=['GET', 'POST'])   # noqa[C901]
@login_required
def transaction(id):

    try:
        transaction = models.db.session.query(
            models.Transaction).filter_by(id=id, user=g.user).one()
    except NoResultFound:
        return redirect(url_for('transactions._transactions'))

    # If this is a child, redirect to the parent
    if transaction.parent_id:
        return redirect(url_for('transactions.transaction',
                                id=transaction.parent_id))

    form = FormTransaction(obj=transaction)
    form_split = None

    form.account.query = models.db.session.query(models.Account) \
        .join(models.Entity, models.Account.entity_id == models.Entity.id) \
        .filter_by(user=g.user) \
        .order_by(models.Entity.name, models.Account.name).all()

    form.bankaccount.query = models.db.session.query(models.BankAccount). \
        filter_by(user=g.user).all()

    # XXX: figure out and document what the fuck is going on here.
    if request.method == 'POST':
        app.logger.debug("Received POST; validate={0}, split={1}".
                         format(form.validate(), request.form.get('split')))

        if request.form.get('split'):
            form_split = _get_form_split(request, form)

        if form.validate():

            if 'delete' in request.form:
                transaction.is_deleted = True
            elif 'undelete' in request.form:
                transaction.is_deleted = False

            new_split = _save_new_split_transaction(request, transaction)

            if request.form.get('update'):

                # Update child accounts.  If a child account is updated
                # a child_account_${id} key will exist within
                # request.form where ${id} is the tx id of the child.
                for child in transaction.children:
                    # Ignore the child if it is the new split
                    if child == new_split:
                        continue
                    child_account_id = request.form.get('child_account_{}'.
                                                        format(child.id))
                    app.logger.debug(
                        'Updating child; {0}, {1}'.
                        format(child.id, child_account_id))

                    # Set child_account_id to the account_id that the
                    # user has set in the form.
                    try:
                        child.account_id = int(child_account_id)
                    except (ValueError, TypeError):
                        # child_account_id is None or NoneType.
                        child.account_id = None
                    finally:
                        app.logger.debug("Adding child {0} to session".
                                         format(child))
                        models.db.session.add(child)

                # Set the amount.
                transaction.set_amount(form.amount.data)

        # Set transaction account based off form.
        account = form.account.data
        if account:
            transaction.account = account
        else:
            transaction.account = None

        if form.attachment.data:
            absfilepath, relfilepath = get_filepath_for_transaction_attachment(
                app.config['TRANSACTION_ATTACHMENTS_UPLOAD_FOLDER'],
                transaction,
                form.attachment.data.filename
            )
            form.attachment.data.save(absfilepath)
            attachment_hash = get_hash_of_file(absfilepath)
            try:
                models.db.session.query(models.TransactionAttachment) \
                    .filter_by(transaction=transaction,
                               attachment_hash=attachment_hash).one()
            except NoResultFound:
                attachment = models.TransactionAttachment(
                    transaction=transaction,
                    filename=form.attachment.data.filename,
                    filepath=relfilepath, attachment_hash=attachment_hash)
                models.db.session.add(attachment)

        if form.note.data:
            try:
                models.db.session.query(models.TransactionNote) \
                    .filter_by(transaction=transaction,
                               note=form.note.data).one()
            except NoResultFound:
                note = models.TransactionNote(transaction=transaction,
                                              note=form.note.data)
                models.db.session.add(note)

        models.db.session.add(transaction)
        models.db.session.commit()

    child_accounts = {}
    for child in transaction.children:
        if child.account:
            child_accounts[child.id] = child.account.id
        app.logger.debug("Children Accounts; {0}".format(child_accounts))

    form.set_defaults(transaction)
    form.set_data(transaction)

    notes = models.db.session.query(models.TransactionNote) \
        .filter_by(transaction=transaction).all()

    attachments = models.db.session.query(models.TransactionAttachment) \
        .filter_by(transaction=transaction).all()

    return render_template('transaction.html', form=form,
                           form_split=form_split, attachments=attachments,
                           transaction=transaction, notes=notes,
                           child_accounts=child_accounts)


@transactions.route('/add', methods=['GET', 'POST'])
@login_required
def add():
    form = FormTransactionAdd()
    form.account.choices = forms.get_account_as_choices()
    form.bankaccount.choices = forms.get_bankaccount_as_choices()

    if form.validate_on_submit():
        transaction = models.Transaction(
            user_id=g.user.id, date=form.date.data, memo=form.memo.data,
            debit=form.get_debit(), credit=form.get_credit(),
            account=form.get_account(), bankaccount=form.get_bankaccount())

        models.db.session.add(transaction)
        models.db.session.commit()

        return redirect(url_for('transactions.transaction',
                                id=transaction.id))

    return render_template('transaction_add.html', form=form)


@transactions.route('/attachment/<int:id>', methods=['GET', 'POST'])
@login_required
def attachment(id):
    """Serve transactionattachment files.

    The attachment filepath is the relative to
    TRANSACTION_ATTACHMENTS_UPLOAD_FOLDER.

    Joining TRANSACTION_ATTACHMENTS_UPLOAD_FOLDER and the
    attachment filepath field will result in the absolute filepath
    of the transaction file.
    """

    try:
        attachment = models.db.session.query(models.TransactionAttachment) \
            .join(models.Transaction,
                  models.TransactionAttachment.transaction_id
                  == models.Transaction.id) \
            .filter(models.TransactionAttachment.id == id,
                    models.Transaction.user_id == g.user.id) \
            .one()
    except NoResultFound:
        abort(404)

    file_dir = os.path.join(
        app.config['TRANSACTION_ATTACHMENTS_UPLOAD_FOLDER'],
        re.sub(
            r"^/",
            "",
            os.path.dirname(
                os.path.realpath(attachment.filepath)
            )
        )
    )
    file_name = os.path.basename(os.path.realpath(attachment.filepath))

    app.logger.info(
        "Serving transaction attachment; file_dir={0}, file_name={1}"
        .format(file_dir, file_name)
    )

    return send_from_directory(file_dir, file_name)


@transactions.route('/import', defaults={'id': None}, methods=['GET', 'POST'])
@transactions.route('/import/<int:id>', methods=['GET', 'POST'])
@login_required
def upload(id):
    form = FormTransactionUpload()

    if form.validate_on_submit():

        filepath = os.path.join(
            app.config['TRANSACTION_UPLOADS_UPLOAD_FOLDER'],
            str(g.user.id), datetime.datetime.now().strftime('%s'),
            secure_filename(form.upload.data.filename))

        filedir = os.path.dirname(os.path.realpath(filepath))

        if not os.path.isdir(filedir):
            os.makedirs(filedir)

        form.upload.data.save(filepath)

        transaction_upload_hash = hashlib.md5(
            open(filepath, 'rb').read()).hexdigest()

        transactionupload = models.TransactionUpload(
            user=g.user, filename=form.upload.data.filename,
            filepath=filepath, upload_hash=transaction_upload_hash)

        models.db.session.add(transactionupload)

        try:
            models.db.session.commit()
        except IntegrityError:
            models.db.session.rollback()
            flash('Failed Uploading Transactions.', 'error')

        # Perform Upload tasks.
        app.logger.info("About to submit import transaction task; "
                        "transactionupload={0}, user={1}".
                        format(transactionupload.id, g.user.id))
        tasks.import_transactions.delay(transactionupload.id, g.user.id)
        app.logger.info("Import transaction task sent; "
                        "transactionupload={0}, user={1}".
                        format(transactionupload.id, g.user.id))

        flash('Uploaded Transactions.', 'success')

    return render_template('transaction_import.html', form=form)
