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
	environment {
		GITHUB_TOKEN = credentials('marianob85-github-jenkins')
	}
	
	stages
	{
		stage('Build package') 
		{
			environment { 
				l_workspace = createWSPath()
			}
			steps
			{
				ws("${env.l_workspace}"){
					dir('ohm'){
						checkout scm
						script {
							env.GITHUB_REPO = sh(script: 'basename $(git remote get-url origin) .git', returnStdout: true).trim()
						}
						sh '''
							export PATH=$PATH:/usr/local/go/bin
							export GOPATH=${l_workspace}
							make package
						'''
						archiveArtifacts artifacts: 'build/**', onlyIfSuccessful: true,  fingerprint: true
						stash includes: 'build/dist/**', name: 'dist'
					}
				}
			}
			post {   
				cleanup {
					cleanWSPath()
				}
			}
		}
		stage('Test') 
		{
			environment { 
				l_workspace = createWSPath()
			}
			steps {
				ws("${env.l_workspace}"){
					dir('ohm'){
						checkout scm
						sh '''
							export GOROOT=/usr/local/go
							export PATH=$PATH:$GOROOT/bin
							export GOPATH=${l_workspace}
							make test
						'''
					}
				}
      		}
			post {   
				cleanup {
					cleanWSPath()
				}
			}
		}
		
		stage('Release') {
			environment { 
				l_workspace = createWSPath()
			}
			when {
				buildingTag()
			}
			steps {
				ws("${env.l_workspace}"){
					dir('ohm'){
						unstash 'dist'
						sh '''
							export GOROOT=/usr/local/go
							export GOPATH=${l_workspace}
							export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
							go get github.com/github-release/github-release
							github-release release --user marianob85 --repo ${GITHUB_REPO} --tag ${TAG_NAME} --name ${TAG_NAME}
							for filename in build/dist/*; do
								[ -e "$filename" ] || continue
								basefilename=$(basename "$filename")
								github-release upload --user marianob85 --repo ${GITHUB_REPO} --tag ${TAG_NAME} --name ${basefilename} --file ${filename}
							done
						'''
					}
				}
			}
			post {   
				cleanup {
					cleanWSPath()
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

def createWSPath() {
	return  "${env.WORKSPACE}/${env.BRANCH_NAME}/${env.STAGE_NAME}/${env.BUILD_NUMBER}".replace('%2F', '_').replace(' ', '_')
}

def cleanWSPath() {
	cleanWs deleteDirs: true, patterns: [[pattern: 'env.l_workspace', type: 'INCLUDE'], [pattern: '', type: 'INCLUDE']]
}