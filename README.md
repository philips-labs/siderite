# siderite
siderite is a companion tool to the iron CLI to make the interaction with the HSDP Iron service more pleasant. It can both prepare the payloads for your tasks and also act as a runner in your dockerized workload to interpret the payload.

# requirements
* IronCLI - https://dev.iron.io/worker/reference/cli/
* CF CLI - https://docs.cloudfoundry.org/cf-cli/install-go-cli.html
* Access to HSDP CF

# configuration
as a first step you need to have a HSDP Iron instances provisioned through the HSDP Iron broker. The service details of this instance should be in your home folder as `~/.iron.json`. This can be done using the sequence of commands shown below:

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
| version | string | version of JSON payload | Y | must be `"1"` for now |
| cmd | []string | command to execute, array string | Y| `["df", "-h"]` |
| env | hashmap | hash with environment variables | N | `{"foo": "bar"}` |


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
```

### 7. encrypt payload.json to payload.enc with CLI tool

```bash
cat payload.json |siderite encrypt > payload.enc
```

### 8. create Docker image and push to private repo

> Dockerfile

```Dockerfile
FROM loafoe/siderite-java8:latest

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
