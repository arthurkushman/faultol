# FaulTol
Fault Tolerance component consumer to save static html files on disk from RabbitMQ

[![Go Report Card](https://goreportcard.com/badge/github.com/arthurkushman/faultol)](https://goreportcard.com/report/github.com/arthurkushman/faultol)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

### Setting ENVs
```bash
FAULTOL_AMQP_CONNECTION=amqp://rabbit:Qwerty11@example.com:5672/
```

```bash
FAULTOL_AMQP_EXCHANGE=fanout_services
```

```bash
FAULTOL_AMQP_EXCHANGE_TYPE=fanout
```

```bash
FAULTOL_AMQP_QUEUE=beta.upload.html_pages_dev
```

```bash
FAULTOL_HTML_PATH=/var/www/shared
```

```bash
FAULTOL_MAX_THREADS=10
```

And a couple of optional settings:

```bash
FAULTOL_AMQP_BINDING_KEY=your_binding_key
```

```bash
FAULTOL_AMQP_CONSUMER_TAG=example_department_tag
```

To set a time to consume you can use:
```bash
FAULTOL_AMQP_LIFETIME=300
``` 
if it has been set to 0 or not set at all consumer will be running forever.

### Message structure

JSON message structure that must be put in RabbitMQ - index.html example:
```json
{
  "uri": "/",
  "data": "<html><head>...</head> ... </html>"
}
```
and for any other pages:
```json
{
  "uri": "/some/route/to/any/123/page",
  "data": "<div> ... </div>"
}
```
or
```json
{
  "uri": "/catalog/cats/basket/94713",
  "data": "<div> ... </div>"
}
```
etc

Those data contents will be converted into paths and put into directories like: `${FAULTOL_HTML_PATH}/some/route/to/any/123/page.html`, `${FAULTOL_HTML_PATH}/catalog/cats/basket/94713.html`

After installation has been completed, in any place of your app - run:
```go
package main

import "faultol"

func main() {
    faultol.Run()
}
```