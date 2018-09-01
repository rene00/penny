pipeline {
    agent any
    environment {
        LANG = 'en_US.UTF-8'
    }
    stages {
        stage('Build') {
            steps {
              sh 'sudo apt install -y make virtualenv gcc libssl-dev dpkg-dev python-dev'
              sh 'make clean'
              sh 'make venv'
            }
        }
        stage('Test') {
            steps {
              sh 'venv/bin/python setup.py test'
            }
        }
    }
}
