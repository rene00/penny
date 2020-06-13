def to_cents(amount):
    return int(round(float(amount) * 100))


def get_credit_debit(amount):
    """Returns a tuple of credit and debit."""
    if amount < 0:
        return (0, amount)
    else:
        return (amount, 0)
