import com.sap.gateway.ip.core.customdev.util.Message
import org.apache.camel.CamelContext
import org.apache.camel.Exchange
import groovy.json.JsonBuilder

def Message processData(Message message) {
    Exchange ex = message.exchange
    CamelContext ctx = ex.getContext()
    String camelVersion = ctx.getVersion().replaceAll(/(\d+.\d+.\d+)-.+/,'$1')

    def builder = new JsonBuilder()
    builder {
        versions {
            'java' System.getProperty('java.version')
            'groovy' GroovySystem.getVersion() 
            'camel' camelVersion
        }
    }

    message.setBody(builder.toString())
    return message
}