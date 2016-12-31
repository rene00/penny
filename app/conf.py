CSRF_ENABLED = True
SECRET_KEY = 'tatokuddMiWradfo'

SQLALCHEMY_DATABASE_URI = 'mysql+pymysql://penny:penny@localhost:3306/penny'
SQLALCHEMY_TRACK_MODIFICATIONS = False

# customizations for login_manager.
SECURITY_CONFIRMABLE=False
SECURITY_LOGIN_USER_TEMPLATE='user/login.html'
SECURITY_PASSWORD_HASH='bcrypt'
SECURITY_PASSWORD_SALT='cecsebWorPenitdiTin'
SECURITY_POST_CONFIRM_VIEW='post_confirm_view'
SECURITY_REGISTERABLE=True
SECURITY_REGISTER_USER_TEMPLATE='user/register.html'
SECURITY_SEND_PASSWORD_CHANGE_EMAIL=False
SECURITY_SEND_PASSWORD_RESET_NOTICE_EMAIL=False
SECURITY_SEND_REGISTER_EMAIL=False

# directory where transaction attachments are stored.
TRANSACTION_ATTACHMENTS_UPLOAD_FOLDER='files/attachments'

# directory where transaction uploads (eg; OFX files) are stored.
TRANSACTION_UPLOADS_UPLOAD_FOLDER='files/uploads'

# RQ
RQ_DEFAULT_URL = 'redis://localhost:6379/0'

# Debug
DEBUG = False

# Logging with Flask-Log
FLASK_LOG_LEVEL = 'INFO'
