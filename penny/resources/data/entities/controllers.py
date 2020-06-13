from penny import models
from flask import (Blueprint, g, jsonify)
from flask_security import login_required

data_entities = Blueprint('data_entities', __name__,
                          url_prefix='/data/entities')


@data_entities.route('/')
@login_required
def entities():
    """Return data on all entities."""
    total = 0
    data = {'rows': []}
    entities = models.db.session.query(models.Entity) \
        .filter_by(user=g.user).all()
    if entities:
        for entity in entities:
            data['rows'].append(entity.dump())
        total = len(entities)
    data['total'] = total
    return jsonify(data)
