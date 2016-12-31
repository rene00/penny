"""empty message

Revision ID: 56bac2207b80
Revises: 19d709c26f6
Create Date: 2016-03-25 04:39:53.286873

"""

# revision identifiers, used by Alembic.
revision = '56bac2207b80'
down_revision = '19d709c26f6'

from alembic import op
import sqlalchemy as sa


def upgrade():
    ### commands auto generated by Alembic - please adjust! ###
    op.add_column('user', sa.Column('active', sa.Boolean(), nullable=True))
    ### end Alembic commands ###


def downgrade():
    ### commands auto generated by Alembic - please adjust! ###
    op.drop_column('user', 'active')
    ### end Alembic commands ###
