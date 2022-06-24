import penny
from penny import models
from sqlalchemy.exc import IntegrityError
import re
from typing import Pattern, Any


def tag_match(user_id: int) -> None:
    app = penny.create_app()
    with app.app_context():
        user = models.User.query.filter_by(id=user_id).one()
        for tag in user.tags:
            for regex in tag.regexes:
                r: Pattern[Any] = re.compile(regex.regex, re.IGNORECASE)
                for transaction in models.Transaction.query.filter(
                    models.Transaction.user_id == user.id,
                    models.Transaction.is_deleted == False,
                    models.Transaction.is_archived == False,
                    ~models.Transaction.tags.any(models.Tag.id.in_([tag.id])),
                ).all():
                    if not r.search(transaction.memo):
                        continue
                    transaction.tags.append(tag)
                    models.db.session.add(transaction)
                    try:
                        models.db.session.commit()
                    except IntegrityError:
                        models.db.session.rollback()
