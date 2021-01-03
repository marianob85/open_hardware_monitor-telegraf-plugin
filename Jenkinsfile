properties(
	[
		buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '5')),
		pipelineTriggers
		(
			[
				pollSCM('0 H(5-6) * * *')
			]
		)
	]
)

pipeline
{
	agent { 
		node { 
			label 'linux && go' 
		} 
	}
	options {
		skipDefaultCheckout true
	}
	
	stages
	{
		stage('Build package') 
		{
			steps
			{
				dir('ohm')
				{
					checkout scm
					sh '''
						export PATH=$PATH:/usr/local/go/bin
						export GOPATH=${WORKSPACE}
						make package
					'''
					archiveArtifacts artifacts: 'build/dist/**', onlyIfSuccessful: true,  fingerprint: true
					//stash includes: 'build/dist/**', name: 'build'
				}
			}
			post {
                always {
					cleanWs()
                }
            }
		}
		stage('Test') 
		{
			steps {
				dir('ohm')
				{
					checkout scm
					sh '''
						export GOROOT=/usr/local/go
						export PATH=$PATH:$GOROOT/bin
						export GOPATH=${WORKSPACE}
						make test
					'''
				}
      		}
			post {
                always {
					cleanWs()
                }
            }
		}
		
		stage ('Release') {
			when {
				buildingTag()
			}

			environment {
				GITHUB_TOKEN = credentials('6347238a-927e-4a50-9f2e-23b50f681fba')
			}

			steps {
				dir('ohm')
				{
					checkout scm
					sh '''
						export GOROOT=/usr/local/go
						export PATH=$PATH:$GOROOT/bin
						export GOPATH=${WORKSPACE}
						goreleaser
					'''
					archiveArtifacts artifacts: 'dist/*', onlyIfSuccessful: true,  fingerprint: true
					//stash includes: 'dist/*', name: 'tag'
				}
			}
			post {
                always {
					cleanWs()
                }
            }
		}
	}
	post 
	{ 
        failure { 
            notifyFailed()
        }
		success { 
            notifySuccessful()
        }
		unstable { 
            notifyFailed()
        }
    }
}

def notifySuccessful() {
	echo 'Sending e-mail'
	mail (to: 'notifier@manobit.com',
         subject: "Job '${env.JOB_NAME}' (${env.BUILD_NUMBER}) success build",
         body: "Please go to ${env.BUILD_URL}.");
}

def notifyFailed() {
	echo 'Sending e-mail'
	mail (to: 'notifier@manobit.com',
         subject: "Job '${env.JOB_NAME}' (${env.BUILD_NUMBER}) failure",
         body: "Please go to ${env.BUILD_URL}.");
}
