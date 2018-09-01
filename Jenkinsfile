pipeline {
    agent any
    environment {
        LANG = 'en_US.UTF-8'
    }
    stages {
        stage('Build') {
            steps {
              sh 'make venv'
            }
        }
    }
}
