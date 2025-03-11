# DEPRECATED: please use https://github.com/dip-software/siderite

# siderite
siderite is a companion tool to the iron CLI to make the interaction with the Iron.io service more pleasant. It can both prepare the payloads for your tasks and also act as a runner in your dockerized workload to interpret the payload.

## Disclaimer

> [!Important]
> This repository is managed as Philips Inner-source / Open-source.
> This repository is NOT endorsed or supported by HSSA&P or I&S Cloud Operations. 
> You are expected to self-support or raise tickets on the Github project and NOT raise tickets in HSP ServiceNow. 

# requirements
* IronCLI - https://dev.iron.io/worker/reference/cli/
* CF CLI - https://docs.cloudfoundry.org/cf-cli/install-go-cli.html
* Access to Cloud foundyr

# install siderite binary
Ensure you have [Go 1.21 or newer](https://golang.org/dl/) installed, then:

```shell
$ go install github.com/philips-labs/siderite@latest
```

# configuration
next you need to have a Iron.io instances provisioned through an Iron.io service broker. The service details of this instance should be in your home folder as `~/.iron.json`. This can be done using the sequence of commands shown below:

```shell
$ cf cs hsdp-iron dev-large-encrypted iron
$ cf csk iron siderite
$ cf service-key iron siderite |tail -n +2 > ~/.iron.json
```

# usage
siderite defines the following JSON payload format
```json
{
	"version": "1",
	"cmd": ["cmd", "-option"],
	"env": {
		"ENV_VARIABLE_NAME": "ENV_VARIABLE_VALUE",
		"FOO": "BAR"
	}
}
```

| field | type |description | required | example      |
|-------|------|-------|----------|--------------|
| version | string | version of JSON payload | Required | must be `"1"` for now |
| cmd | []string | command to execute, array string | Required | `["df", "-h"]` |
| env | hashmap | hash with environment variables | Optional | `{"foo": "bar"}` |

## logging

The siderite binary supports direct logging to HSDP logging when the following environment variables
are set:

| environment | description | required |
|-------------|-------------|----------|
| SIDERITE_LOGINGESTOR_PRODUCT_KEY| The HSDP logging product key | Required |
| SIDERITE_LOGINGESTOR_KEY | The HSDP logging shared key | Optional |
| SIDERITE_LOGINGESTOR_SECRET | The HSDP logging shared secret | Optional |
| SIDERITE_LOGINGESTOR_URL | The HSDP logging base URL | Required when not setting region and environment |
| SIDERITE_LOGINGESTOR_SERVICE_ID | The HSDP service identity ID to use | Optional |
| SIDERITE_LOGINGESTOR_SERVICE_PRIVATE_KEY | The private key belonging to the service identity | Optional |
| SIDERITE_LOGINGESTOR_REGION | The HSDP region | Required for service identity |
| SIDERITE_LOGINGESTOR_ENVIRONMENT | The HSDP environment (`client-test`, `prod`) | Required for service identity |

## logging using HSDP Logdrainer

If you only have access to a Logdrainer endpoint URL then you can configure it as well

| environment | description | required |
|-------------|-------------|----------|
| SIDERITE_LOGDRAINER_URL | The logdrainer endpoint used in CF | Optional |

# commands

## doctor
checks your system for correct configuration and suggest steps to take
```shell
$ siderite doctor
[✓] iron CLI installed (version 0.1.6)
[✓] iron configuration file (/Users/andy/.iron.json)
[✓] cf CLI installed (cf version 6.49.0+d0dfa93bb.2020-01-07)
```

## encrypt
encrypts input (stdin by default) with the cluster public key
```shell
$ echo '{"cmd":"ls"}'|siderite encrypt
VRUYw6MZqakMz1KX6Ag21EfwEj9VBCV0jVpo3buEY8kIqaZK+dgC7YoJNjQ7tFfM9bPFMw+8yVawNG0u4IeLeSkSH+aLCA8bXVMl5hKVVOelY+eGceD9qXhTq9RDAyuY2RJ3XCHIUfQre1XIn8jO2GCtIUSIvKJ7XB6lYPg2jocXsYQ8xvVOnESiWexTur94afdB82HpFx6yDcHlrblovEdqtVk/fzOZ8A==
```

## env2payload
converts ENV style input (on stdin by default) to the siderite JSON payload format
```shell
$ echo 'FOO=BAR'|siderite env2payload -c "echo","\$FOO"
{
  "version": "1",
  "env": {
    "FOO": "BAR"
  },
  "cmd": [
    "echo",
    "$FOO"
  ]
}
```

## runner
opens the payload file references by `PAYLOAD_FILE` environment and executes the command, mapping all output to stdout. This mode should be used as the `ENTRYPOINT` command in your Docker image

## task and function

The siderite binary also acts as the entrypoint / command for [hsdp_function](https://registry.terraform.io/providers/philips-software/hsdp/latest/docs/guides/functions) compatible Docker images. 

# example usage

## converting a CF Java8 app to an IronIO scheduled task
  
The example steps below assume that your CF app is deployed under name `app` and your application is available as `app.jar`. For best results your `app.jar` should have a "run once" mode where the processing starts immediately after startup and terminates once done. This ensures you only consume the time your app is run instead of having your task terminated by IronIO after the 1 hour default timeout.

### 1. install iron CLI

```bash
curl -sSL http://get.iron.io/cli | sh
```
Further details: https://dev.iron.io/worker/reference/cli/

### 2. provision Iron instance via marketplace

```bash
cf cs hsdp-iron dev-large-encrypted iron
```

### 3. setup service key

```bash
cf csk iron siderite
```

### 4. setup ~/.iron.json
```bash
cf service-key iron siderite | tail -n +2 > ~/.iron.json
```

### 5. capture ENV variables from your existing CF app

```bash
cf ssh app -c env | siderite env2payload > payload.json
```

### 6. add cmd to payload.json

```json
{
  "version": "1",
  "cmd": ["java", "-jar", "/data/app.jar"],
  "env": {
    "VCAP_SERVICES": "[REDACTED]",
    "VCAP_APPLICATION": "[REDACTED]",
    "ADD_MORE_STUFF": "here"
  }
}
```

### 7. encrypt payload.json to payload.enc with CLI tool

```bash
cat payload.json |siderite encrypt > payload.enc
```

### 8. create Docker image and push to private repo

:triangular_flag_on_post: Below is an example only. Do not use Java8, it is obsolete! 

> [Dockerfile](https://github.com/philips-labs/siderite-java8)

```Dockerfile
FROM loafoe/siderite-java8:v0.11.0

RUN mkdir -p /data
ADD app.jar /data
```

```bash
docker login docker.na1.hsdp.io
docker build -t docker.na1.hsdp.io/yournamespace/app .
docker push docker.na1.hsdp.io/yournamespace/app
```

> The `loafoe/siderite-java8:latest` contains the latest Java 8 runtime and the `siderite` tool as the `ENTRYPOINT`. It will detect the decrypted `payload.json`, set the ENVironment according to the `env` content and will execute the `cmd` command in the container

### 9. register docker image as code with Iron 

> Store docker credentials with Iron

```bash
iron docker login \
  -url https://docker.na1.hsdp.io \
  -u ServiceUserName \
  -p ServidePassword \
  -e your.name@philips.com
```

```bash
iron register docker.na1.hsdp.io/yournamespace/app
```

### 10. get the cluster ID

```bash
cf service-key iron key| \
  grep -v Getting| \
  jq .cluster_info[].cluster_id -r
```

### 11. schedule your task using Iron.io CLI

> Below example schedules your app code to run every hour. Make sure `payload.enc` file and `cluster_id` value are available!

```bash
iron worker schedule \
  -cluster replace_with_cluser_id \
  -run-every 3600 \
  -payload-file payload.enc \
  docker.na1.hsdp.io/yournamespace/app
```

# best practices

- Package your workload in Docker images
- Encrypt payload data
- Limit log output to actionable log entries only
- Use [logproxy](https://github.com/philips-software/logproxy) to forward IronIO logs to HSDP logging.
  
# siderite name
Siderite is a mineral composed of iron(II) carbonate (FeCO3). It takes its name from the Greek word σίδηρος sideros, "iron". It is a valuable iron mineral, since it is 48% iron and contains no sulfur or phosphorus. [Wikipedia](https://en.wikipedia.org/wiki/Siderite)

# license

License is MIT
