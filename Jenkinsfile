pipeline {
     agent any
         environment {
             AWS_ACCOUNT_ID="${env.AWS_ACCOUNT_ID}"
             AWS_DEFAULT_REGION="${env.AWS_DEFAULT_REGION}"
             DOCKER_BUILDKIT="1"
             IMAGE_REPO_NAME="${GIT_URL.tokenize('/.')[-2]}"
             BUILD_NUMBER=sh(script: "git rev-parse --short HEAD", returnStdout: true)
             EC2_CREDENTIAL_ID="${env.EC2_CREDENTIAL_ID}"
             EC2_USER="${EC2_USER}"
             TAG="${IMAGE_REPO_NAME}:${BUILD_NUMBER}"
             ECR_URL="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_DEFAULT_REGION}.amazonaws.com"
         }
     options {
         skipStagesAfterUnstable()
     }
     stages {
          stage('Environment variables'){
            steps {
                script {
                    echo "AWS_ACCOUNT_ID = ${AWS_ACCOUNT_ID}"
                    echo "AWS_DEFAULT_REGION = ${AWS_DEFAULT_REGION}"
                    echo "DOCKER_BUILDKIT = ${DOCKER_BUILDKIT}"
                    echo "IMAGE_REPO_NAME = ${IMAGE_REPO_NAME}"
                    echo "BUILD_NUMBER = ${BUILD_NUMBER}"
                    echo "TAG = ${TAG}"
                    echo "EC2_CREDENTIAL_ID = ${EC2_CREDENTIAL_ID}"
                }
            }
          }

          stage('Clone repository') {
             steps {
                 script{
                 checkout scm
                 }
             }
         }

         stage('Build') {
             steps {
                 script{
                  app = docker.build("${IMAGE_REPO_NAME}")
                 }
             }
         }
         stage('Deploy') {
             steps {
                 script{
                     docker.withRegistry("https://${ECR_URL}", "ecr:${AWS_DEFAULT_REGION}:aws-credentials") {
                         app.push()
                         app.push("release")
                     }
                 }
             }
         }

         stage('Run container on prod'){
            steps {
                sshagent(["${EC2_CREDENTIAL_ID}"]) {
                     sh "ssh -o StrictHostKeyChecking=no ${EC2_USER} 'deploy_prod.sh'"
                }
            }
         }
     }
         post {
             always {
                 cleanWs()
             }
         }
 }