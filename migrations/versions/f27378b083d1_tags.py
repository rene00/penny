"""tags

Revision ID: f27378b083d1
Revises: e16799a209b1
Create Date: 2022-06-17 07:00:38.481092

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = 'f27378b083d1'
down_revision = 'e16799a209b1'
branch_labels = None
depends_on = None


def upgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.create_table('tag_match_filter_regex',
    sa.Column('id', sa.Integer(), nullable=False),
    sa.Column('regex', sa.String(length=128), nullable=True),
    sa.Column('tag_id', sa.Integer(), nullable=True),
    sa.Column('date_added', sa.DateTime(), nullable=True),
    sa.ForeignKeyConstraint(['tag_id'], ['tag.id'], ),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('regex', 'tag_id')
    )
    # ### end Alembic commands ###


def downgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.drop_table('tag_match_filter_regex')
    # ### end Alembic commands ###
