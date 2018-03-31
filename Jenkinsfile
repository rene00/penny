pipeline {
    agent {
        docker {
            image 'ubuntu:xenial'
            args '--user root:sudo'
        }
    }
    environment {
        LANG = 'en_US.UTF-8'
    }
    stages {
        stage('setup') {
            steps {
                sh "echo 'Acquire::http::Proxy \"${HTTP_PROXY}\";' > /etc/apt/apt.conf"
                sh 'apt-get update -y'
                sh 'apt-get upgrade -y'
                sh 'apt-get install -y gcc python3-dev tox locales'
                sh "echo 'en_US.UTF-8 UTF-8' > /etc/locale.gen"
                sh '/usr/sbin/locale-gen'
            }
        }
        stage('test') {
            steps {
                sh 'tox'
            }
        }
    }
}
