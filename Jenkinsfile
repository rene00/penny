pipeline {
    agent any
    environment {
        LANG = 'en_US.UTF-8'
    }
    stages {
        stage('Build') {
            steps {
              sh 'sudo apt install -y make'
              sh 'make venv'
            }
        }
    }
}
