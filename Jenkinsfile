@Library('github.com/releaseworks/jenkinslib') _
pipeline {
     agent any
         environment {
             AWS_ACCOUNT_ID="${env.AWS_ACCOUNT_ID}"
             AWS_DEFAULT_REGION="${env.AWS_DEFAULT_REGION}"
             DOCKER_BUILDKIT="1"
             SERVICE_NAME="${GIT_URL.tokenize('/.')[-2]}"
             SHORT_COMMIT=sh(script: "git rev-parse --short HEAD", returnStdout: true).trim()
             EC2_CREDENTIAL_ID="${env.EC2_CREDENTIAL_ID}"
             EC2_USER="${EC2_USER}"
             ECR_URL="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_DEFAULT_REGION}.amazonaws.com"
             REPOSITORY_URI="${ECR_URL}/${SERVICE_NAME}"

             // ECS properties

             TASK_FAMILY="${SERVICE_NAME}" // at least one container needs to have the same name as the task definition
             DESIRED_COUNT="1"
             CLUSTER_NAME = "hof_dev_cluster"
             AWS_ACCESS_KEY_ID = "${env.AWS_ID_USR}"
             AWS_SECRET_ACCESS_KEY = "${env.AWS_ID_PSW}"
             EXECUTION_ROLE_ARN = "${env.EXECUTION_ROLE_ARN}"
         }

     stages {
          stage('Environment variables'){
            steps {
                script {
                    echo "AWS_DEFAULT_REGION = ${AWS_DEFAULT_REGION}"
                    echo "DOCKER_BUILDKIT = ${DOCKER_BUILDKIT}"
                    echo "SERVICE_NAME = ${SERVICE_NAME}"
                    echo "SHORT_COMMIT = ${SHORT_COMMIT}"
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
                  docker.build("${REPOSITORY_URI}", ".")
                 }
             }
         }
         stage('Deploy') {
             steps {
                 script{
                     docker.withRegistry("https://${ECR_URL}", "ecr:${AWS_DEFAULT_REGION}:aws-credentials") {
                       docker.image("${REPOSITORY_URI}").push("${SHORT_COMMIT}")
                     }
                 }
             }
         }

        stage('Deploy Image to ECS') {
            steps{
                // prepare task definition file
                sh """sed -e "s;%REPOSITORY_URI%;${REPOSITORY_URI};g" -e "s;%SHORT_COMMIT%;${SHORT_COMMIT};g" -e "s;%TASK_FAMILY%;${TASK_FAMILY};g" -e "s;%SERVICE_NAME%;${SERVICE_NAME};g" -e "s;%EXECUTION_ROLE_ARN%;${EXECUTION_ROLE_ARN};g" -e "s;%AWS_ENDPOINT%;${env.OLD_S3_AWS_ENDPOINT};g" -e "s;%AWS_SECRET%;${env.OLD_S3_AWS_SECRET};g" -e "s;%AWS_REGION%;${env.OLD_S3_AWS_REGION};g" -e "s;%JWT_SECRET%;${env.JWT_SECRET};g" -e "s;%PAYSTACK_SECRET%;${env.PAYSTACK_SECRET};g" -e "s;%DATABASE_URL%;${env.DATABASE_URL};g" taskdef_template.json > taskdef_${SHORT_COMMIT}.json"""
                script {
                    // Register task definition
                    AWS("ecs register-task-definition --output json --cli-input-json file://${WORKSPACE}/taskdef_${SHORT_COMMIT}.json > ${env.WORKSPACE}/temp.json")
                    def projects = readJSON file: "${env.WORKSPACE}/temp.json"
                    def TASK_REVISION = projects.taskDefinition.revision

                    // update service
                    AWS("ecs update-service --cluster ${CLUSTER_NAME} --service ${SERVICE_NAME} --task-definition ${TASK_FAMILY}:${TASK_REVISION} --desired-count ${DESIRED_COUNT}")
                }
            }
        }
        stage('Remove docker image') {
            steps{
                // Remove images
                sh "docker rmi $REPOSITORY_URI"
            }
        }
       }
         post {
             always {
                 cleanWs()
             }
         }
 }