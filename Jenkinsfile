pipeline {
  agent {
    node {
      label 'master'
    }

  }
  stages {
    stage('Checkout') {
      steps {
        checkout scm
      }
    }
    stage('Build Image') {
      steps {
        script {
          sh '''
            LOGIN=$(vault kv get -format=json secret/jenkins/dockerhub | jq .data.data)
            USER=$(echo $LOGIN | jq -r .user)
            PASS=$(echo $LOGIN | jq -r .password)
            docker login --username $USER --password $PASS
          '''

          sh 'docker build -t lnd:$BRANCH_NAME-$BUILD_NUMBER .'
          sh 'docker tag lnd:$BRANCH_NAME-$BUILD_NUMBER lightningpeach/lnd:$BRANCH_NAME-$BUILD_NUMBER'
          sh 'docker tag lnd:$BRANCH_NAME-$BUILD_NUMBER lightningpeach/lnd:latest'
        }
      }
    }
      stage('Push Image') {
        steps {
          sh 'docker push lightningpeach/lnd:$BRANCH_NAME-$BUILD_NUMBER'
          sh 'docker push lightningpeach/lnd:latest'
        }
      }
    }
  }
