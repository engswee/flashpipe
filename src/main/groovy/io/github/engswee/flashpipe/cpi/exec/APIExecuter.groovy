package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken
import org.slf4j.Logger
import org.slf4j.LoggerFactory

abstract class APIExecuter {
    final HTTPExecuter httpExecuter
    final String oauthTokenHost

    static Logger logger = LoggerFactory.getLogger(APIExecuter)
    
    APIExecuter() {
        this.oauthTokenHost = System.getenv('HOST_OAUTH') ?: null

        def tenantMgmtHost = getMandatoryEnvVar('HOST_TMN')
        if (this.oauthTokenHost) {
            def oauthClientID = getMandatoryEnvVar('OAUTH_CLIENTID')
            def oauthClientSecret = getMandatoryEnvVar('OAUTH_CLIENTSECRET')
            def oauthTokenPath = System.getenv('HOST_OAUTH_PATH') ?: null
            String oauthToken = OAuthToken.get('https', this.oauthTokenHost, 443, oauthClientID, oauthClientSecret, oauthTokenPath)
            this.httpExecuter = HTTPExecuterApacheImpl.newInstance('https', tenantMgmtHost, 443, oauthToken)
        } else {
            def basicAuthUser = getMandatoryEnvVar('BASIC_USERID')
            def basicAuthPassword = getMandatoryEnvVar('BASIC_PASSWORD')
            this.httpExecuter = HTTPExecuterApacheImpl.newInstance('https', tenantMgmtHost, 443, basicAuthUser, basicAuthPassword)
        }
    }

    protected String getMandatoryEnvVar(String envVarName) {
        def envVar = System.getenv(envVarName)
        if (!envVar) {
            logger.error("🛑 Mandatory environment variable ${envVarName} not populated")
            System.exit(1)
        }
        return envVar
    }
    
    protected void validateInputContainsNoSecrets(String envVarName) {
        String envVar = System.getenv(envVarName)
        if (envVar) {
            checkSecretNotInVariable(envVarName, 'OAUTH_CLIENTID')
            checkSecretNotInVariable(envVarName, 'OAUTH_CLIENTSECRET')
            checkSecretNotInVariable(envVarName, 'BASIC_USERID')
            checkSecretNotInVariable(envVarName, 'BASIC_PASSWORD')
        }
    }
    
    protected void checkSecretNotInVariable(String inputVarName, String secretVarName) {
        String inputValue = System.getenv(inputVarName)
        String secretValue = System.getenv(secretVarName)
        if (secretValue && inputValue.contains(secretValue)) {
            logger.error("🛑 Security risk! Environment variable ${inputVarName} contains value of secret variable ${secretVarName}")
            System.exit(1)
        }
    }

    abstract void getEnvironmentVariables()

    abstract void execute()
}
