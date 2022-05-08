"""empty message

Revision ID: 87246e047222
Revises: 983cac6d79fb
Create Date: 2022-04-17 16:40:40.072133

"""

# revision identifiers, used by Alembic.
revision = '87246e047222'
down_revision = '983cac6d79fb'

from alembic import op
import sqlalchemy as sa


def upgrade():
    op.add_column('user', sa.Column('fs_uniquifier', sa.String(length=255), nullable=False))
    with op.batch_alter_table("user") as batch_op:
        batch_op.create_unique_constraint('fs_uniquifier_unique', ['fs_uniquifier'])


def downgrade():
    op.drop_column('user', 'fs_uniquifier')
