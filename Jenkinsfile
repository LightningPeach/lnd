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
    
    stage('Buld Image') {
      steps {
        script {
          def secrets = [
              [$class: 'VaultSecret', path: 'secret/jenkins/dockerhub', secretValues: [
                  [$class: 'VaultSecretValue', envVar: 'duser', vaultKey: 'user'],
                  [$class: 'VaultSecretValue', envVar: 'dpass', vaultKey: 'password']]]
          ]

          wrap([$class: 'VaultBuildWrapper', vaultSecrets: secrets]) {
            sh 'docker login --username $duser --password $dpass'
          }

          sh 'docker build -t lightningpeach/lnd:$BRANCH_NAME-$BUILD_NUMBER .'
          sh 'docker tag lightningpeach/lnd:$BRANCH_NAME-$BUILD_NUMBER lightningpeach/lnd:$BRANCH_NAME-$BUILD_NUMBER'        
          sh 'docker tag lightningpeach/lnd:$BRANCH_NAME-$BUILD_NUMBER lightningpeach/lnd:latest'
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
