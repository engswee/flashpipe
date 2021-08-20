import com.sap.gateway.ip.core.customdev.util.Message
import groovy.json.JsonBuilder
import groovy.json.JsonSlurper
import java.time.LocalDate
import java.time.format.DateTimeFormatter

def Message processData(Message message) {
  Reader reader = message.getBody(Reader)
  def input = new JsonSlurper().parse(reader)

  def builder = new JsonBuilder()
  builder.PurchaseOrder {
    'Header' {
      'ID' input.Order.Header.OrderNumber
      'DocumentDate' LocalDate.parse(input.Order.Header.Date, DateTimeFormatter.ofPattern('yyyyMMdd')).format(DateTimeFormatter.ofPattern('yyyy-MM-dd'))
    }
    def items = input.Order.Items.findAll { item -> item.Valid }
    'Items' items.collect { item ->
      [
          'ItemNumber' : item.ItemNumber.padLeft(3, '0'),
          'ProductCode': item.MaterialNumber,
          'Quantity'   : item.Quantity
      ]
    }
  }

  message.setBody(builder.toPrettyString())
  return message
}