package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken

abstract class APIExecuter {
    final HTTPExecuter httpExecuter
    final String oauthTokenHost

    APIExecuter() {
        this.oauthTokenHost = System.getenv('HOST_OAUTH') ?: null

        def tenantMgmtHost = getMandatoryEnvVar('HOST_TMN')
        if (this.oauthTokenHost) {
            def oauthClientID = getMandatoryEnvVar('OAUTH_CLIENTID')
            def oauthClientSecret = getMandatoryEnvVar('OAUTH_CLIENTSECRET')
            String oauthToken = OAuthToken.get('https', this.oauthTokenHost, 443, oauthClientID, oauthClientSecret)
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
            println "[ERROR] - Mandatory environment variable ${envVarName} not populated"
            System.exit(1)
        }
        return envVar
    }

    abstract void execute()
}
