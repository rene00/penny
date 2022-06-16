from penny import models
from penny.common import forms
from penny.resources.accountmatches.util import (
    update_filters,
    add_filter,
    update_details,
)
from flask import Blueprint, g, render_template, url_for, redirect, request
from flask_security.decorators import auth_required
from sqlalchemy.orm.exc import NoResultFound
from penny.resources.accountmatches.forms import (
    FormAccountMatch,
    FormAccountMatchFilter,
)


accountmatches = Blueprint("accountmatches", __name__)


@accountmatches.route("/accountmatches")
@auth_required()
def _accountmatches():
    return render_template(
        "accountmatches.html", data_url=url_for("data_accountmatches.accountmatches")
    )


@accountmatches.route("/accountmatches/<int:id>", methods=["GET", "POST"])
@auth_required()
def accountmatch(id):

    try:
        accountmatch = (
            models.db.session.query(models.AccountMatch)
            .filter_by(id=id, user=g.user)
            .one()
        )
    except NoResultFound:
        return url_for("accountmatches._accountmatches")

    form = FormAccountMatch(obj=accountmatch)
    form.account.choices = forms.get_account_as_choices()
    form.bankaccount.choices = forms.get_bankaccount_as_choices()
    filters = (
        models.db.session.query(models.AccountMatchFilterRegex)
        .filter_by(accountmatch=accountmatch)
        .order_by(models.AccountMatchFilterRegex.date_added.desc())
    )

    if form.validate_on_submit():
        accountmatch.name = form.name.data
        accountmatch.desc = form.desc.data

        if form.account.data:
            accountmatch.account_id = form.account.data

        if form.bankaccount.data:
            accountmatch.bankaccount_id = form.bankaccount.data

        models.db.session.add(accountmatch)

        # Add any new filters.
        if request.form.get("regex"):
            accountmatchfilterregex = models.AccountMatchFilterRegex(
                regex=request.form.get("regex"), accountmatch=accountmatch
            )
            models.db.session.add(accountmatchfilterregex)

        models.db.session.commit()

        if "update" in request.form:
            # Add any new fitlers.
            new_filter = add_filter(accountmatch, request)

            # Update existing filters.
            update_filters(accountmatch, request, new_filter)

            # Update other details.
            accountmatch = update_details(accountmatch, form)

        elif "filter_add" in request.form:
            return render_template(
                "accountmatch.html",
                filters=filters,
                form=form.set_defaults(accountmatch),
                accountmatch=accountmatch,
                form_filter=FormAccountMatchFilter(),
            )

    form.set_defaults(accountmatch)

    return render_template(
        "accountmatch.html", form=form, filters=filters, accountmatch=accountmatch
    )


@accountmatches.route("/accountmatches/add", methods=["GET", "POST"])
@auth_required()
def add():
    form = FormAccountMatch()
    form_filter = FormAccountMatchFilter()
    form.account.choices = forms.get_account_as_choices()
    form.bankaccount.choices = forms.get_bankaccount_as_choices()

    if form.validate_on_submit():
        accountmatch = models.AccountMatch(user_id=g.user.id)
        accountmatch.name = form.name.data
        accountmatch.desc = form.desc.data

        if form.account.data:
            accountmatch.account_id = form.account.data

        if form.bankaccount.data:
            accountmatch.bankaccount_id = form.bankaccount.data

        if form_filter and form_filter.regex.data:
            # accountmatchfilterregex
            amfr = models.AccountMatchFilterRegex(
                accountmatch=accountmatch, regex=form_filter.regex.data
            )
            models.db.session.add(amfr)

        models.db.session.add(accountmatch)
        models.db.session.commit()

        return redirect(url_for("accountmatches.accountmatch", id=accountmatch.id))

    return render_template("accountmatch.html", form=form, form_filter=form_filter)
