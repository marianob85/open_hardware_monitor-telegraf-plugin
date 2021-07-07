properties(
	[
		buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '5')),
	]
)
pipeline
{
	agent any
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
			agent{ label "linux/u18.04/go:1.15.13" }
			steps
			{
				checkout scm
				script {
					env.GITHUB_REPO = sh(script: 'basename $(git remote get-url origin) .git', returnStdout: true).trim()
				}
				sh '''
					make package
				'''
				archiveArtifacts artifacts: 'build/**', onlyIfSuccessful: true,  fingerprint: true
				stash includes: 'build/dist/**', name: 'dist'
			}
		}
		stage('Test') 
		{
			agent{ label "linux/u18.04/go:1.15.13" }
			steps {
				checkout scm
				sh '''
					make test
				'''
      		}
		}
		
		stage('Release') {
			when {
				buildingTag()
			}
			agent{ label "linux/u18.04/go:1.15.13" }
			steps {
				unstash 'dist'
				sh '''
					export GOPATH=${PWD}
					go get github.com/github-release/github-release
					bin/github-release release --user marianob85 --repo ${GITHUB_REPO} --tag ${TAG_NAME} --name ${TAG_NAME}
					for filename in build/dist/*; do
						[ -e "$filename" ] || continue
						basefilename=$(basename "$filename")
						bin/github-release upload --user marianob85 --repo ${GITHUB_REPO} --tag ${TAG_NAME} --name ${basefilename} --file ${filename}
					done
				'''
			}
		}
	}
	post { 
        changed { 
            emailext body: 'Please go to ${env.BUILD_URL}', to: '${DEFAULT_RECIPIENTS}', subject: "Job ${env.JOB_NAME} (${env.BUILD_NUMBER}) ${currentBuild.currentResult}".replaceAll("%2F", "/")
        }
    }
}