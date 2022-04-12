"""empty message

Revision ID: 983cac6d79fb
Revises: f490e985f5c9
Create Date: 2022-04-13 08:10:55.846098

"""

# revision identifiers, used by Alembic.
revision = '983cac6d79fb'
down_revision = 'f490e985f5c9'

from alembic import op
import sqlalchemy as sa


def upgrade():
    op.rename_table('transaction', 'tx')


def downgrade():
    op.rename_table('tx', 'transaction')
